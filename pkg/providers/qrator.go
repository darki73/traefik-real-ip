package providers

import (
	"net"
	"net/http"
	"strings"
)

const (
	_qratorProviderXQratorIPSourceHeader = "X-Qrator-IP-Source"
)

// QratorProvider is the provider for Qrator.
type QratorProvider struct {
	name              string
	headers           []string
	values            map[string]string
	excludedNetworks  []*net.IPNet
	excludedAddresses []net.IP
}

// Initialize initializes the provider.
func (qp *QratorProvider) Initialize(excludedNetworks []*net.IPNet, excludedAddresses []net.IP) ProviderInterface {
	return &QratorProvider{
		name: "qrator",
		headers: []string{
			_qratorProviderXQratorIPSourceHeader,
		},
		values:            map[string]string{},
		excludedNetworks:  excludedNetworks,
		excludedAddresses: excludedAddresses,
	}
}

// GetName returns the name of the provider.
func (qp *QratorProvider) GetName() string {
	return qp.name
}

// GetHeaders returns the headers which are specific to this provider.
func (qp *QratorProvider) GetHeaders() []string {
	return qp.headers
}

// GetValues returns the header => value pairs which are specific to this provider.
func (qp *QratorProvider) GetValues() map[string]string {
	return qp.values
}

// GetRealIP returns the real IP address of the client.
func (qp *QratorProvider) GetRealIP(request *http.Request) string {
	qp.fillValues(request)

	for _, value := range qp.GetValues() {
		if !qp.isExcludedIP(value) {
			return value
		}
	}

	return ""
}

// fillValues fills the values map with the headers from the request.
func (qp *QratorProvider) fillValues(request *http.Request) {
	for _, header := range qp.GetHeaders() {
		if value := request.Header.Get(header); value != "" {
			qp.values[header] = strings.TrimSpace(value)
		}
	}
}

// getExcludedNetworks returns the list of excluded networks.
func (qp *QratorProvider) getExcludedNetworks() []*net.IPNet {
	return qp.excludedNetworks
}

// getExcludedAddresses returns the list of excluded addresses.
func (qp *QratorProvider) getExcludedAddresses() []net.IP {
	return qp.excludedAddresses
}

// isExcludedIP returns true if the IP is excluded.
func (qp *QratorProvider) isExcludedIP(address string) bool {
	ip := net.ParseIP(address)

	if ip == nil {
		return true
	}

	for _, excludedNetwork := range qp.getExcludedNetworks() {
		if excludedNetwork.Contains(ip) {
			return true
		}
	}

	for _, excludedAddress := range qp.getExcludedAddresses() {
		if ip.Equal(excludedAddress) {
			return true
		}
	}

	return false
}
