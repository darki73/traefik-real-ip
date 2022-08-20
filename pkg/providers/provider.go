package providers

import (
	"errors"
	"net"
	"net/http"
)

var (
	ErrorInvalidProvider          = errors.New("invalid provider")
	ErrorInvalidPreferredProvider = errors.New("invalid preferred provider")
)

// ProviderInterface is the interface that all providers must implement.
type ProviderInterface interface {
	// Initialize initializes the provider.
	Initialize(excludedNetworks []*net.IPNet, excludedAddresses []net.IP) ProviderInterface
	// GetName returns the name of the provider.
	GetName() string
	// GetHeaders returns the headers which are specific to this provider.
	GetHeaders() []string
	// GetValues returns the header => value pairs which are specific to this provider.
	GetValues() map[string]string
	// GetRealIP returns the real IP address of the client.
	GetRealIP(request *http.Request) string
	// fillValues fills the values map with the headers from the request.
	fillValues(request *http.Request)
	// getExcludedNetworks returns the list of excluded networks.
	getExcludedNetworks() []*net.IPNet
	// getExcludedAddresses returns the list of excluded addresses.
	getExcludedAddresses() []net.IP
	// isExcludedIP returns true if the IP is excluded.
	isExcludedIP(address string) bool
}

// Providers is the list of providers.
type Providers struct {
	providers map[string]ProviderInterface
}

// Initialize initializes the providers.
func Initialize(excludedNetworks []*net.IPNet, excludedAddresses []net.IP) *Providers {
	instance := &Providers{
		providers: map[string]ProviderInterface{
			"generic":    &GenericProvider{},
			"cloudflare": &CloudflareProvider{},
			"qrator":     &QratorProvider{},
		},
	}

	for index, provider := range instance.providers {
		instance.providers[index] = provider.Initialize(excludedNetworks, excludedAddresses)
	}

	return instance
}

// GetAvailableProviders returns the list of available providers.
func (providers *Providers) GetAvailableProviders() []string {
	var providersList []string

	for index := range providers.providers {
		providersList = append(providersList, index)
	}

	return providersList
}

// IsValidProvider returns true if the provider is valid.
func (providers *Providers) IsValidProvider(provider string) bool {
	for _, value := range providers.GetAvailableProviders() {
		if value == provider {
			return true
		}
	}
	return false
}

// GetProvider returns the provider with the given name.
func (providers *Providers) GetProvider(provider string) ProviderInterface {
	return providers.providers[provider]
}

// GetProviders returns the list of all providers.
func (providers *Providers) GetProviders() map[string]ProviderInterface {
	return providers.providers
}
