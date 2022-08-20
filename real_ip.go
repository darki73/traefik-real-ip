package traefik_real_ip

import (
	"context"
	"github.com/darki73/traefik-real-ip/pkg/providers"
	"net"
	"net/http"
	"strings"
)

// Config holds configuration passed to the plugin.
type Config struct {
	ExcludedNetworks  []string `json:"excludedNetworks,omitempty" toml:"excludedNetworks,omitempty" yaml:"excludedNetworks,omitempty"`
	ExcludedAddresses []string `json:"excludedAddresses,omitempty" toml:"excludedAddresses,omitempty" yaml:"excludedAddresses,omitempty"`
	Providers         []string `json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty"`
	PreferredProvider string   `json:"preferredProvider,omitempty" toml:"preferredProvider,omitempty" yaml:"preferredProvider,omitempty"`
}

// CreateConfig creates the default plugin configuration if no parameters are passed.
func CreateConfig() *Config {
	return &Config{
		ExcludedNetworks:  []string{},
		ExcludedAddresses: []string{},
		Providers:         []string{},
		PreferredProvider: "",
	}
}

// TraefikRealIP holds the necessary components of a Traefik plugin.
type TraefikRealIP struct {
	next              http.Handler
	name              string
	excludedNetworks  []*net.IPNet
	excludedAddresses []net.IP
	providers         map[string]providers.ProviderInterface
	preferredProvider string
	providersIPs      map[string]string
}

// New instantiates and returns the required components used to handle HTTP request.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	trip := &TraefikRealIP{
		next:              next,
		name:              name,
		providers:         map[string]providers.ProviderInterface{},
		preferredProvider: config.PreferredProvider,
		providersIPs:      map[string]string{},
	}

	for _, value := range config.ExcludedNetworks {
		_, excludedNetwork, err := net.ParseCIDR(value)
		if err != nil {
			return nil, err
		}
		trip.excludedNetworks = append(trip.excludedNetworks, excludedNetwork)
	}

	for _, value := range config.ExcludedAddresses {
		ip := net.ParseIP(value)

		if ip != nil {
			trip.excludedAddresses = append(trip.excludedAddresses, ip)
		}
	}

	manager := providers.Initialize(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())

	if config.PreferredProvider != "" {
		if !manager.IsValidProvider(config.PreferredProvider) {
			return nil, providers.ErrorInvalidPreferredProvider
		}
	}

	if len(config.Providers) == 0 {
		trip.providers = manager.GetProviders()
	} else {
		for _, value := range config.Providers {
			provider := strings.TrimSpace(strings.ToLower(value))
			if manager.IsValidProvider(provider) {
				trip.providers[provider] = manager.GetProvider(provider)
			} else {
				return nil, providers.ErrorInvalidProvider
			}
		}

		if _, ok := trip.providers["generic"]; !ok {
			trip.providers["generic"] = manager.GetProvider("generic")
		}
	}

	return trip, nil
}

// ServeHTTP handles the HTTP request.
func (trip *TraefikRealIP) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	for providerName, providerInstance := range trip.GetProviders() {
		realIP := providerInstance.GetRealIP(request)
		if realIP != "" {
			trip.providersIPs[providerName] = realIP
		}
	}

	realIP := ""

	if _, ok := trip.providersIPs["generic"]; ok {
		realIP = trip.providersIPs["generic"]
		delete(trip.providersIPs, "generic")
	}

	if len(trip.providersIPs) > 0 {
		if trip.HasPreferredProvider() {
			if _, ok := trip.providersIPs[trip.GetPreferredProvider()]; ok {
				realIP = trip.providersIPs[trip.GetPreferredProvider()]
			}
		} else {
			for _, provider := range trip.GetProviders() {
				if _, ok := trip.providersIPs[provider.GetName()]; ok {
					realIP = trip.providersIPs[provider.GetName()]
					break
				}
			}
		}
	}

	if realIP != "" {
		request.Header.Set("X-Forwarded-For", realIP)
		request.Header.Set("X-Real-Ip", realIP)
	}

	trip.next.ServeHTTP(responseWriter, request)
}

// GetExcludedNetworks returns list of excluded networks.
func (trip *TraefikRealIP) GetExcludedNetworks() []*net.IPNet {
	return trip.excludedNetworks
}

// GetExcludedAddresses returns list of excluded addresses.
func (trip *TraefikRealIP) GetExcludedAddresses() []net.IP {
	return trip.excludedAddresses
}

// GetProviders returns list of providers.
func (trip *TraefikRealIP) GetProviders() map[string]providers.ProviderInterface {
	return trip.providers
}

// GetPreferredProvider returns preferred provider.
func (trip *TraefikRealIP) GetPreferredProvider() string {
	return trip.preferredProvider
}

// HasPreferredProvider returns true if preferred provider is set.
func (trip *TraefikRealIP) HasPreferredProvider() bool {
	return trip.preferredProvider != ""
}
