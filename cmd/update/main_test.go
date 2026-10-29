package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestValidateGeneratedDocumentation(t *testing.T) {
	if err := validateGeneratedDocumentation(3, false, 3, 2, 2, true); err != nil {
		t.Fatalf("complete documentation rejected: %v", err)
	}

	tests := []struct {
		name                 string
		componentCount       int
		componentFailed      bool
		detailedCount        int
		guideCount           int
		colorsGenerated      bool
		wantMessageSubstring string
	}{
		{"no components", 0, false, 0, 2, true, "no component"},
		{"component failure", 3, true, 3, 2, true, "component files"},
		{"missing detail", 3, false, 2, 2, true, "2 of 3 detailed"},
		{"missing guide", 3, false, 3, 1, true, "1 of 2 guide"},
		{"missing colors", 3, false, 3, 2, false, "colors.md"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateGeneratedDocumentation(test.componentCount, test.componentFailed, test.detailedCount, test.guideCount, 2, test.colorsGenerated)
			if err == nil || !strings.Contains(err.Error(), test.wantMessageSubstring) {
				t.Fatalf("error = %v, want substring %q", err, test.wantMessageSubstring)
			}
		})
	}
}

func TestPinnedReleaseSources(t *testing.T) {
	version, err := parseVersion("5.7.50\r\n")
	if err != nil || version != "5.7.50" {
		t.Fatalf("parseVersion() = %q, %v", version, err)
	}
	for _, invalid := range []string{"", "latest", "master", "v5.7.50", "5.7", "5.07.50", "5.7.50-beta.1", "6.0.0", "../../master"} {
		if _, err := parseVersion(invalid); err == nil {
			t.Errorf("parseVersion(%q) succeeded", invalid)
		}
	}
	if got := sourceURL(version, "skills/daisyui/SKILL.md"); got != "https://raw.githubusercontent.com/saadeghi/daisyui/refs/tags/v5.7.50/skills/daisyui/SKILL.md" {
		t.Fatalf("sourceURL() = %q", got)
	}
	index := "[components](./components/)\n[Button](./components/button.md)\n[File input](./components/file-input.md)\n"
	names, err := componentNamesFromIndex(index)
	if err != nil || !reflect.DeepEqual(names, []string{"button", "file-input"}) {
		t.Fatalf("componentNamesFromIndex() = %v, %v", names, err)
	}
	for _, invalid := range []string{"no links", index + "[Button again](./components/button.md)", "[Bad](./components/../outside.md)", "[Bad](./components/button?query.md)", "[Bad](./components/Button.md)"} {
		if _, err := componentNamesFromIndex(invalid); err == nil {
			t.Errorf("componentNamesFromIndex(%q) succeeded", invalid)
		}
	}
	wantColors := "## daisyUI 5 colors\n\n### Names\n- primary\n"
	colors, err := extractColors([]byte("---\nname: colors\n---\n\n" + wantColors))
	if err != nil || string(colors) != wantColors {
		t.Fatalf("extractColors() = %q, %v", colors, err)
	}
	if _, err := extractColors([]byte("404 page")); err == nil {
		t.Fatal("extractColors accepted a response without color documentation")
	}
}

func TestDownloadDocumentsBoundedConcurrency(t *testing.T) {
	const total = downloadWorkers * 2
	started := make(chan struct{}, total)
	release := make(chan struct{})
	var releaseOnce sync.Once
	unblock := func() { releaseOnce.Do(func() { close(release) }) }
	var active, peak atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); current > old; old = peak.Load() {
			if peak.CompareAndSwap(old, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		if r.URL.Path == "/failed" {
			http.Error(w, "unavailable", http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, r.URL.Path)
	}))
	t.Cleanup(server.Close)
	t.Cleanup(unblock)
	dir := t.TempDir()
	documents := make([]documentDownload, total)
	for i := range documents {
		name := fmt.Sprintf("document-%d", i)
		documents[i] = documentDownload{name, server.URL + "/" + name}
	}
	documents[total-1].url = server.URL + "/failed"
	done := make(chan int, 1)
	go func() { done <- downloadDocuments(server.Client(), dir, documents) }()
	for range downloadWorkers {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("downloads did not run concurrently")
		}
	}
	unblock()
	select {
	case count := <-done:
		if count != total-1 {
			t.Fatalf("downloaded %d documents, want %d", count, total-1)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("downloads did not complete")
	}
	if got := peak.Load(); got != downloadWorkers {
		t.Errorf("peak concurrent downloads = %d, want %d", got, downloadWorkers)
	}
	for _, doc := range documents[:total-1] {
		data, err := os.ReadFile(filepath.Join(dir, doc.name+".md"))
		if err != nil || string(data) != "/"+doc.name {
			t.Errorf("downloaded %s = %q, %v", doc.name, data, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, documents[total-1].name+".md")); !os.IsNotExist(err) {
		t.Errorf("failed response was saved: %v", err)
	}
	if count := downloadDocuments(server.Client(), dir, nil); count != 0 {
		t.Errorf("empty batch count = %d", count)
	}
}

func TestReplaceGeneratedDocumentation(t *testing.T) {
	names := []string{"components", "docs", "guide", "colors.md"}
	writeFixture := func(t *testing.T, root, name, value string) {
		t.Helper()
		path := filepath.Join(root, name)
		if name != "colors.md" {
			path = filepath.Join(path, "example.md")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, existing := range []bool{false, true} {
		for failAt := -1; failAt < len(names); failAt++ {
			t.Run(fmt.Sprintf("existing=%t/failAt=%d", existing, failAt), func(t *testing.T) {
				dst, src := t.TempDir(), t.TempDir()
				for i, name := range names {
					if existing {
						writeFixture(t, dst, name, "old "+name)
					}
					if i != failAt {
						writeFixture(t, src, name, "new "+name)
					}
				}
				err := replaceGeneratedDocumentation(dst, src)
				if failAt == -1 && err != nil || failAt != -1 && err == nil {
					t.Fatalf("replaceGeneratedDocumentation() = %v", err)
				}
				for _, name := range names {
					if failAt != -1 && !existing {
						if _, err := os.Stat(filepath.Join(dst, name)); !os.IsNotExist(err) {
							t.Errorf("rollback left %s behind: %v", name, err)
						}
						continue
					}
					path := filepath.Join(dst, name)
					if name != "colors.md" {
						path = filepath.Join(path, "example.md")
					}
					want := "old " + name
					if failAt == -1 {
						want = "new " + name
					}
					data, err := os.ReadFile(path)
					if err != nil || string(data) != want {
						t.Errorf("%s = %q, %v; want %q", path, data, err, want)
					}
				}
				backups, err := filepath.Glob(filepath.Join(dst, ".*.bak-*"))
				if err != nil || len(backups) != 0 {
					t.Errorf("backups remain: %v, %v", backups, err)
				}
			})
		}
	}
}

func TestReadLimitedAndDownloadValidation(t *testing.T) {
	if _, err := readLimited(strings.NewReader("123456"), 5); err == nil {
		t.Fatal("readLimited accepted an oversized response")
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = response.Write([]byte("  \n"))
	}))
	t.Cleanup(server.Close)
	if _, err := downloadFile(server.Client(), server.URL); err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("downloadFile empty response error = %v, want empty-body error", err)
	}
}

func TestComponentNameFromSection(t *testing.T) {
	tests := []struct {
		name    string
		section string
		want    string
		wantErr bool
	}{
		{
			name:    "title words become hyphenated slug",
			section: "### File input\n[File input documentation](https://daisyui.com/components/file-input/)\n#### Class names\n",
			want:    "file-input",
		},
		{
			name:    "canonical slug can reorder title words",
			section: "### Browser mockup\n[Browser mockup documentation](https://daisyui.com/components/mockup-browser/)\n#### Syntax\n",
			want:    "mockup-browser",
		},
		{
			name:    "fragment is accepted",
			section: "[Theme controller documentation](https://daisyui.com/components/theme-controller/#example)",
			want:    "theme-controller",
		},
		{
			name:    "untrusted host is rejected",
			section: "[Button documentation](https://example.com/components/button/)",
			wantErr: true,
		},
		{
			name:    "unsafe path is rejected",
			section: "[Button documentation](https://daisyui.com/components/../button/)",
			wantErr: true,
		},
		{
			name:    "missing link is rejected",
			section: "### Button\n#### Class names\n",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := componentNameFromSection(test.section)
			if test.wantErr {
				if err == nil {
					t.Fatalf("componentNameFromSection() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("componentNameFromSection() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("componentNameFromSection() = %q, want %q", got, test.want)
			}
		})
	}
}
