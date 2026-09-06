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
