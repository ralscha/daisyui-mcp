package theme

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/lucasb-eyer/go-colorful"
	_ "golang.org/x/image/webp"
)

var imageHTTPClient = newImageHTTPClient()

const (
	maxImageBytes  = 10 << 20
	maxImagePixels = 16_000_000
)

var blockedImageNetworks = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),     // this network
	netip.MustParsePrefix("100.64.0.0/10"), // shared address space
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),  // documentation
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),    // reserved for future use
	netip.MustParsePrefix("64:ff9b::/96"),   // IPv4/IPv6 translation
	netip.MustParsePrefix("64:ff9b:1::/48"), // local-use IPv4/IPv6 translation
	netip.MustParsePrefix("100::/64"),       // discard-only addresses
	netip.MustParsePrefix("2001::/32"),      // Teredo
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"), // 6to4
	netip.MustParsePrefix("fec0::/10"), // deprecated site-local addresses
}

func newImageHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = dialPublicAddress
	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return validateRemoteImageURL(request.URL)
		},
	}
}

func validateRemoteImageURL(imageURL *url.URL) error {
	if imageURL == nil {
		return fmt.Errorf("image URL must use http or https")
	}
	scheme := strings.ToLower(imageURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("image URL must use http or https")
	}
	if imageURL.Hostname() == "" {
		return fmt.Errorf("image URL must include a host")
	}
	if imageURL.User != nil {
		return fmt.Errorf("image URL must not include credentials")
	}
	return nil
}

func dialPublicAddress(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("invalid image host address: %w", err)
	}

	addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve image host %q: %w", host, err)
	}
	publicAddresses := make([]net.IPAddr, 0, len(addresses))
	for _, candidate := range addresses {
		if isPublicImageAddress(candidate.IP) {
			publicAddresses = append(publicAddresses, candidate)
		}
	}
	if len(publicAddresses) == 0 {
		return nil, fmt.Errorf("image host %q resolves only to private or non-public addresses", host)
	}

	dialer := &net.Dialer{}
	var lastErr error
	for _, candidate := range publicAddresses {
		connection, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		if dialErr == nil {
			return connection, nil
		}
		lastErr = dialErr
	}
	return nil, fmt.Errorf("connect to image host %q: %w", host, lastErr)
}

func isPublicImageAddress(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
		return false
	}
	for _, network := range blockedImageNetworks {
		if network.Contains(address) {
			return false
		}
	}
	return true
}

func ExtractThemeFromImage(imagePathOrURL string) (ThemeInput, error) {
	return ExtractThemeFromImageContext(context.Background(), imagePathOrURL)
}

func ExtractThemeFromImageContext(ctx context.Context, imagePathOrURL string) (ThemeInput, error) {
	if err := ctx.Err(); err != nil {
		return ThemeInput{}, err
	}
	var data []byte

	trimmedSource := strings.TrimSpace(imagePathOrURL)
	lowerSource := strings.ToLower(trimmedSource)
	if strings.HasPrefix(lowerSource, "http://") || strings.HasPrefix(lowerSource, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimmedSource, nil)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to create image request: %w", err)
		}
		if err := validateRemoteImageURL(req.URL); err != nil {
			return ThemeInput{}, err
		}
		req.Header.Set("Accept", "image/*")

		resp, err := imageHTTPClient.Do(req)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return ThemeInput{}, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}
		if resp.ContentLength > maxImageBytes {
			return ThemeInput{}, fmt.Errorf("image is larger than %d bytes", maxImageBytes)
		}

		data, err = readLimited(resp.Body, maxImageBytes)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to read image from URL: %w", err)
		}
	} else {
		if strings.Contains(trimmedSource, "://") {
			return ThemeInput{}, fmt.Errorf("image URL must use http or https")
		}
		file, err := os.Open(imagePathOrURL)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to open local image file: %w", err)
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to inspect local image: %w", err)
		}
		if !info.Mode().IsRegular() {
			return ThemeInput{}, fmt.Errorf("local image must be a regular file")
		}
		if info.Size() > maxImageBytes {
			return ThemeInput{}, fmt.Errorf("image is larger than %d bytes", maxImageBytes)
		}

		data, err = readLimited(file, maxImageBytes)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to read local image: %w", err)
		}
	}

	if err := ctx.Err(); err != nil {
		return ThemeInput{}, err
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to decode image metadata: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > maxImagePixels/config.Height {
		return ThemeInput{}, fmt.Errorf("image dimensions %dx%d exceed the limit of %d pixels", config.Width, config.Height, maxImagePixels)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to decode image: %w", err)
	}

	return extractPalette(ctx, img)
}

// Normally sample at most 128 by 128 pixels, including for wide or tall images.
// If every sample is transparent, scan for up to 16,384 visible pixels instead
// so sparse strokes between sample points still produce a palette.
// Five-bit RGB buckets merge nearby colors; alpha-weighted averages preserve
// the source colors without allowing invisible pixels to skew the palette.
func extractPalette(ctx context.Context, img image.Image) (ThemeInput, error) {
	type bucket struct {
		key                      uint16
		red, green, blue, weight uint64
	}
	bounds := img.Bounds()
	stepX := max(1, (bounds.Dx()+127)/128)
	stepY := max(1, (bounds.Dy()+127)/128)
	buckets := make(map[uint16]bucket)
	addPixel := func(x, y int) bool {
		pixel := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
		if pixel.A == 0 {
			return false
		}
		key := uint16(pixel.R>>3)<<10 | uint16(pixel.G>>3)<<5 | uint16(pixel.B>>3)
		entry := buckets[key]
		entry.key = key
		alpha := uint64(pixel.A)
		entry.red += uint64(pixel.R) * alpha
		entry.green += uint64(pixel.G) * alpha
		entry.blue += uint64(pixel.B) * alpha
		entry.weight += alpha
		buckets[key] = entry
		return true
	}
	for y := bounds.Min.Y + stepY/2; y < bounds.Max.Y; y += stepY {
		if err := ctx.Err(); err != nil {
			return ThemeInput{}, err
		}
		for x := bounds.Min.X + stepX/2; x < bounds.Max.X; x += stepX {
			addPixel(x, y)
		}
	}
	if len(buckets) == 0 {
		visible := 0
	scan:
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			if err := ctx.Err(); err != nil {
				return ThemeInput{}, err
			}
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				if (x-bounds.Min.X)%1024 == 0 {
					if err := ctx.Err(); err != nil {
						return ThemeInput{}, err
					}
				}
				if addPixel(x, y) {
					visible++
					if visible == 128*128 {
						break scan
					}
				}
			}
		}
	}
	if len(buckets) == 0 {
		return ThemeInput{}, fmt.Errorf("image has no visible colors")
	}
	ordered := make([]bucket, 0, len(buckets))
	for _, entry := range buckets {
		ordered = append(ordered, entry)
	}
	slices.SortFunc(ordered, func(a, b bucket) int {
		if a.weight > b.weight {
			return -1
		}
		if a.weight < b.weight {
			return 1
		}
		return int(a.key) - int(b.key)
	})
	type swatch struct {
		color                             colorful.Color
		saturation, lightness, population float64
	}
	swatches := make([]swatch, len(ordered))
	for i, entry := range ordered {
		scale := float64(entry.weight) * 255
		c := colorful.Color{R: float64(entry.red) / scale, G: float64(entry.green) / scale, B: float64(entry.blue) / scale}
		_, saturation, lightness := c.Hsl()
		swatches[i] = swatch{c, saturation, lightness, float64(entry.weight) / float64(ordered[0].weight)}
	}
	selected := make(map[int]bool)
	choose := func(saturation, lightness float64, distinct bool) string {
		best, bestScore := 0, -1.0
		for i, candidate := range swatches {
			if distinct && selected[i] {
				continue
			}
			score := 2*(1-math.Abs(candidate.saturation-saturation)) + 1 - math.Abs(candidate.lightness-lightness) + candidate.population
			if score > bestScore {
				best, bestScore = i, score
			}
		}
		selected[best] = true
		return swatches[best].color.Hex()
	}
	return ThemeInput{
		Primary:   choose(1, 0.5, true),
		Secondary: choose(0.8, 0.7, true),
		Accent:    choose(1, 0.3, true),
		Neutral:   choose(0, 0.2, false),
	}, nil
}

func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("image is larger than %d bytes", maxBytes)
	}
	return data, nil
}
