package traefik_real_ip

import (
	"context"
	"fmt"
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
	next               http.Handler
	name               string
	excludedNetworks   []*net.IPNet
	excludedAddresses  []net.IP
	availableProviders []string
	genericProvider    *providers.GenericProvider
	cloudflareProvider *providers.CloudflareProvider
	qratorProvider     *providers.QratorProvider
	preferredProvider  string
	providersIPs       map[string]string
}

// New instantiates and returns the required components used to handle HTTP request.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	trip := &TraefikRealIP{
		next:               next,
		name:               name,
		availableProviders: []string{"generic", "cloudflare", "qrator"},
		preferredProvider:  config.PreferredProvider,
		providersIPs:       make(map[string]string),
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

	if config.PreferredProvider != "" {
		if !trip.IsValidProvider(config.PreferredProvider) {
			return nil, fmt.Errorf(
				"preferred provider %s is not valid, only the following ones are supported: %s",
				config.PreferredProvider,
				strings.Join(trip.availableProviders, ", "),
			)
		}
	}

	trip.genericProvider = providers.InitializeGenericProvider(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())

	for _, provider := range config.Providers {
		if !trip.IsValidProvider(provider) {
			return nil, fmt.Errorf("provider %s is not valid, only the following ones are supported: %s", provider, strings.Join(trip.availableProviders, ", "))
		}
	}

	if config.Providers != nil || len(config.Providers) == 0 {
		trip.cloudflareProvider = providers.InitializeCloudflareProvider(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())
		trip.qratorProvider = providers.InitializeQratorProvider(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())
	} else {
		if trip.ConfigHasProvider("cloudflare", config.Providers) {
			trip.cloudflareProvider = providers.InitializeCloudflareProvider(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())
		}

		if trip.ConfigHasProvider("qrator", config.Providers) {
			trip.qratorProvider = providers.InitializeQratorProvider(trip.GetExcludedNetworks(), trip.GetExcludedAddresses())
		}
	}

	return trip, nil
}

// ServeHTTP handles the HTTP request.
func (trip *TraefikRealIP) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	realIP := ""

	if trip.HasPreferredProvider() {
		if trip.GetPreferredProvider() == "cloudflare" {
			realIP = trip.cloudflareProvider.GetRealIP(request)
		}
		if trip.GetPreferredProvider() == "qrator" {
			realIP = trip.qratorProvider.GetRealIP(request)
		}
	}

	if realIP == "" {
		realIP = trip.genericProvider.GetRealIP(request)
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

// GetPreferredProvider returns preferred provider.
func (trip *TraefikRealIP) GetPreferredProvider() string {
	return trip.preferredProvider
}

// HasPreferredProvider returns true if preferred provider is set.
func (trip *TraefikRealIP) HasPreferredProvider() bool {
	return trip.preferredProvider != ""
}

// IsValidProvider returns true if provider is valid.
func (trip *TraefikRealIP) IsValidProvider(provider string) bool {
	for _, value := range trip.availableProviders {
		if value == provider {
			return true
		}
	}
	return false
}

// ConfigHasProvider returns true if provider is configured.
func (trip *TraefikRealIP) ConfigHasProvider(provider string, providers []string) bool {
	for _, value := range providers {
		if value == provider {
			return true
		}
	}
	return false
}
