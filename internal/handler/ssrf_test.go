package handler

import (
	"testing"

	"spendwise-ms/internal/dto"
)

func TestSSRF_RejectsLoopbackAndPrivateIPs(t *testing.T) {
	invalidURLs := []string{
		"http://127.0.0.1/admin",
		"http://localhost:8080/internal",
		"http://[::1]/status",
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.1/secret",
		"http://192.168.1.1/router",
		"ftp://example.com/file",
	}

	for _, u := range invalidURLs {
		err := dto.IsSafeURL(u)
		if err == nil {
			t.Errorf("Expected SSRF protection to reject '%s', but got nil", u)
		}
	}
}

func TestSSRF_AcceptsPublicSafeURLs(t *testing.T) {
	validURLs := []string{
		"https://api.spendwise.com/v1/data",
		"https://example.com/webhook",
	}

	for _, u := range validURLs {
		err := dto.IsSafeURL(u)
		if err != nil {
			t.Errorf("Expected '%s' to be accepted as safe URL, got error: %v", u, err)
		}
	}
}
