package providers

import (
	"net"
	"net/http"
	"strings"
)

const (
	_genericProviderXForwardedForHeader = "X-Forwarded-For"
	_genericProviderXRealIPHeader       = "X-Real-Ip"
)

// GenericProvider is the generic provider.
type GenericProvider struct {
	name              string
	headers           []string
	values            map[string]string
	excludedNetworks  []*net.IPNet
	excludedAddresses []net.IP
}

// InitializeGenericProvider initializes the Generic provider.
func InitializeGenericProvider(excludedNetworks []*net.IPNet, excludedAddresses []net.IP) *GenericProvider {
	return &GenericProvider{
		name: "generic",
		headers: []string{
			_genericProviderXForwardedForHeader,
			_genericProviderXRealIPHeader,
		},
		values:            map[string]string{},
		excludedNetworks:  excludedNetworks,
		excludedAddresses: excludedAddresses,
	}
}

// GetName returns the name of the provider.
func (gp *GenericProvider) GetName() string {
	return gp.name
}

// GetHeaders returns the headers which are specific to this provider.
func (gp *GenericProvider) GetHeaders() []string {
	return gp.headers
}

// GetValues returns the header => value pairs which are specific to this provider.
func (gp *GenericProvider) GetValues() map[string]string {
	return gp.values
}

// GetRealIP returns the real IP address of the client.
func (gp *GenericProvider) GetRealIP(request *http.Request) string {
	gp.fillValues(request)

	if value, ok := gp.GetValues()[_genericProviderXRealIPHeader]; ok {
		if !gp.isExcludedIP(value) {
			return value
		}
	}

	if value, ok := gp.GetValues()[_genericProviderXForwardedForHeader]; ok {
		forwardChain := strings.Split(value, ",")
		for index, ip := range forwardChain {
			forwardChain[index] = strings.TrimSpace(ip)
		}
		for _, ip := range forwardChain {
			if !gp.isExcludedIP(ip) {
				return ip
			}
		}
	}

	return ""
}

// fillValues fills the values map with the headers from the request.
func (gp *GenericProvider) fillValues(request *http.Request) {
	for _, header := range gp.GetHeaders() {
		if value := request.Header.Get(header); value != "" {
			gp.values[header] = strings.TrimSpace(value)
		}
	}
}

// getExcludedNetworks returns the list of excluded networks.
func (gp *GenericProvider) getExcludedNetworks() []*net.IPNet {
	return gp.excludedNetworks
}

// getExcludedAddresses returns the list of excluded addresses.
func (gp *GenericProvider) getExcludedAddresses() []net.IP {
	return gp.excludedAddresses
}

// isExcludedIP returns true if the IP is excluded.
func (gp *GenericProvider) isExcludedIP(address string) bool {
	ip := net.ParseIP(address)

	if ip == nil {
		return true
	}

	for _, excludedNetwork := range gp.getExcludedNetworks() {
		if excludedNetwork.Contains(ip) {
			return true
		}
	}

	for _, excludedAddress := range gp.getExcludedAddresses() {
		if ip.Equal(excludedAddress) {
			return true
		}
	}

	return false
}
