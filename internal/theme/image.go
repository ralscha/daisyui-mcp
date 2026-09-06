package theme

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/generaltso/vibrant"
	_ "golang.org/x/image/webp"
)

var imageHTTPClient = newImageHTTPClient()

const (
	maxImageBytes  = 10 << 20
	maxImagePixels = 16_000_000
)

var blockedImageNetworks = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"), // shared address space
	netip.MustParsePrefix("192.0.0.0/24"),  // IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),  // documentation
	netip.MustParsePrefix("198.18.0.0/15"), // benchmarking
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("64:ff9b::/96"), // IPv4/IPv6 translation
	netip.MustParsePrefix("2001::/32"),    // Teredo
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

		data, err = readLimited(file, maxImageBytes)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to read local image: %w", err)
		}
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

	palette, err := vibrant.NewPaletteFromImage(img)
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to generate color palette: %w", err)
	}

	swatches := palette.ExtractAwesome()
	themeInput := ThemeInput{}

	if swatch, ok := swatches["Vibrant"]; ok && swatch != nil {
		themeInput.Primary = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["LightVibrant"]; ok && swatch != nil {
		themeInput.Secondary = swatch.Color.RGBHex()
	} else if swatch, ok := swatches["Muted"]; ok && swatch != nil {
		themeInput.Secondary = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["DarkVibrant"]; ok && swatch != nil {
		themeInput.Accent = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["DarkMuted"]; ok && swatch != nil {
		themeInput.Neutral = swatch.Color.RGBHex()
	}

	return themeInput, nil
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
