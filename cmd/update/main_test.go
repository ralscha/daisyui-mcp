package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
