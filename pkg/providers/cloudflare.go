package providers

import (
	"net"
	"net/http"
	"strings"
)

const (
	_cloudflareProviderTrueClientIPHeader   = "True-Client-IP"
	_cloudflareProviderCFConnectingIPHeader = "CF-Connecting-IP"
)

// CloudflareProvider is the provider for Cloudflare.
type CloudflareProvider struct {
	name              string
	headers           []string
	values            map[string]string
	excludedNetworks  []*net.IPNet
	excludedAddresses []net.IP
}

// Initialize initializes the provider.
func (cfp *CloudflareProvider) Initialize(excludedNetworks []*net.IPNet, excludedAddresses []net.IP) ProviderInterface {
	return &CloudflareProvider{
		name: "cloudflare",
		headers: []string{
			_cloudflareProviderTrueClientIPHeader,
			_cloudflareProviderCFConnectingIPHeader,
		},
		values:            map[string]string{},
		excludedNetworks:  excludedNetworks,
		excludedAddresses: excludedAddresses,
	}
}

// GetName returns the name of the provider.
func (cfp *CloudflareProvider) GetName() string {
	return cfp.name
}

// GetHeaders returns the headers which are specific to this provider.
func (cfp *CloudflareProvider) GetHeaders() []string {
	return cfp.headers
}

// GetValues returns the header => value pairs which are specific to this provider.
func (cfp *CloudflareProvider) GetValues() map[string]string {
	return cfp.values
}

// GetRealIP returns the real IP address of the client.
func (cfp *CloudflareProvider) GetRealIP(request *http.Request) string {
	cfp.fillValues(request)

	for _, value := range cfp.GetValues() {
		if !cfp.isExcludedIP(value) {
			return value
		}
	}

	return ""
}

// fillValues fills the values map with the headers from the request.
func (cfp *CloudflareProvider) fillValues(request *http.Request) {
	for _, header := range cfp.GetHeaders() {
		if value := request.Header.Get(header); value != "" {
			cfp.values[header] = strings.TrimSpace(value)
		}
	}
}

// getExcludedNetworks returns the list of excluded networks.
func (cfp *CloudflareProvider) getExcludedNetworks() []*net.IPNet {
	return cfp.excludedNetworks
}

// getExcludedAddresses returns the list of excluded addresses.
func (cfp *CloudflareProvider) getExcludedAddresses() []net.IP {
	return cfp.excludedAddresses
}

// isExcludedIP returns true if the IP is excluded.
func (cfp *CloudflareProvider) isExcludedIP(address string) bool {
	ip := net.ParseIP(address)

	if ip == nil {
		return true
	}

	for _, excludedNetwork := range cfp.getExcludedNetworks() {
		if excludedNetwork.Contains(ip) {
			return true
		}
	}

	for _, excludedAddress := range cfp.getExcludedAddresses() {
		if ip.Equal(excludedAddress) {
			return true
		}
	}

	return false
}
