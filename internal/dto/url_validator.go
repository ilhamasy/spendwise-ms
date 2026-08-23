package dto

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

var (
	ErrInvalidScheme   = errors.New("url scheme must be http or https")
	ErrLoopbackAddress = errors.New("loopback or localhost addresses are not allowed")
	ErrPrivateIP       = errors.New("private or internal IP addresses are not allowed")
	ErrCloudMetadata   = errors.New("cloud metadata service endpoint is not allowed")
)

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.64.0.0/10",
	}
	for _, cidr := range privateRanges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err == nil && subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func IsSafeURL(rawURL string) error {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return errors.New("empty url")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return errors.New("invalid url format")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return ErrInvalidScheme
	}

	host := parsed.Hostname()
	if host == "" {
		return errors.New("missing host in url")
	}

	lowerHost := strings.ToLower(host)
	if lowerHost == "localhost" || lowerHost == "127.0.0.1" || lowerHost == "::1" {
		return ErrLoopbackAddress
	}

	if lowerHost == "169.254.169.254" || strings.Contains(lowerHost, "metadata") {
		return ErrCloudMetadata
	}

	ip := net.ParseIP(host)
	if ip != nil && isPrivateIP(ip) {
		return ErrPrivateIP
	}

	return nil
}
