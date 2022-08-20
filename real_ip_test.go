package traefik_real_ip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTraefikRealIP(framework *testing.T) {
	testCases := []struct {
		description   string
		config        *Config
		expectedError bool
		inputHeaders  map[string]string
		expectedIP    string
	}{
		{
			description: "CreateConfig should return a default configuration if no parameters are passed.",
			config:      &Config{},
		},
		{
			description:   "CreateConfig should return an error if an invalid excluded network is passed.",
			config:        &Config{ExcludedNetworks: []string{"invalid"}},
			expectedError: true,
		},
		{
			description: "CreateConfig should return an error if an invalid excluded address is passed.",
			config:      &Config{ExcludedAddresses: []string{"invalid"}},
		},
		{
			description:   "CreateConfig should return an error if an invalid preferred provider is passed.",
			config:        &Config{PreferredProvider: "invalid"},
			expectedError: true,
		},
		{
			description:   "CreateConfig should return an error if an invalid provider is passed.",
			config:        &Config{Providers: []string{"invalid"}},
			expectedError: true,
		},
		{
			description: "CreateConfig should not throw an error if a valid preferred provider is passed.",
			config:      &Config{PreferredProvider: "qrator"},
		},
		{
			description: "CreateConfig should not throw an error if a valid provider is passed.",
			config:      &Config{Providers: []string{"qrator"}},
		},
		{
			description: "CreateConfig should not throw an error if a valid excluded network is passed.",
			config:      &Config{ExcludedNetworks: []string{"0.0.0.0/0"}},
		},
		{
			description: "CreateConfig should not throw an error if a valid excluded address is passed.",
			config:      &Config{ExcludedAddresses: []string{"192.168.1.1"}},
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should not be present",
			config:      &Config{ExcludedNetworks: []string{"127.0.0.1/24"}},
			inputHeaders: map[string]string{
				"X-Forwarded-For": "127.0.0.2",
			},
			expectedIP: "",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present for generic provider",
			config:      &Config{},
			inputHeaders: map[string]string{
				"X-Forwarded-For": "10.0.0.20",
				"X-Real-Ip":       "10.0.0.20",
			},
			expectedIP: "10.0.0.20",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present and Qrator provider result should be preferred",
			config:      &Config{PreferredProvider: "qrator"},
			inputHeaders: map[string]string{
				"X-Forwarded-For":    "10.0.0.20",
				"X-Real-Ip":          "10.0.0.20",
				"X-Qrator-IP-Source": "10.0.0.30",
				"True-Client-IP":     "10.0.0.40",
				"CF-Connecting-IP":   "10.0.0.40",
			},
			expectedIP: "10.0.0.30",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present and Cloudflare provider result should be preferred",
			config:      &Config{PreferredProvider: "cloudflare"},
			inputHeaders: map[string]string{
				"X-Forwarded-For":    "10.0.0.20",
				"X-Real-Ip":          "10.0.0.20",
				"X-Qrator-IP-Source": "10.0.0.30",
				"True-Client-IP":     "10.0.0.40",
				"CF-Connecting-IP":   "10.0.0.40",
			},
			expectedIP: "10.0.0.40",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present and Generic provider result should be preferred",
			config:      &Config{PreferredProvider: "generic"},
			inputHeaders: map[string]string{
				"X-Forwarded-For":    "10.0.0.20",
				"X-Real-Ip":          "10.0.0.20",
				"X-Qrator-IP-Source": "10.0.0.30",
				"True-Client-IP":     "10.0.0.40",
				"CF-Connecting-IP":   "10.0.0.40",
			},
			expectedIP: "10.0.0.20",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present, Qrator is preferred provider, but Generic should be returned",
			config:      &Config{PreferredProvider: "qrator"},
			inputHeaders: map[string]string{
				"X-Forwarded-For":  "10.0.0.20",
				"X-Real-Ip":        "10.0.0.20",
				"True-Client-IP":   "10.0.0.40",
				"CF-Connecting-IP": "10.0.0.40",
			},
			expectedIP: "10.0.0.20",
		},
		{
			description: "X-Real-Ip or X-Forwarded-For headers should be present, Cloudflare is preferred provider, but Generic should be returned",
			config:      &Config{PreferredProvider: "cloudflare"},
			inputHeaders: map[string]string{
				"X-Forwarded-For":    "10.0.0.20",
				"X-Real-Ip":          "10.0.0.20",
				"X-Qrator-IP-Source": "10.0.0.30",
			},
			expectedIP: "10.0.0.20",
		},
	}

	for _, test := range testCases {
		test := test
		framework.Run(test.description, func(framework *testing.T) {
			framework.Parallel()

			next := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {})
			trip, err := New(context.Background(), next, test.config, "traefik-real-ip")

			if test.expectedError {
				assert.Error(framework, err)
			} else {
				require.NoError(framework, err)
				assert.NotNil(framework, trip)

				if test.inputHeaders != nil && len(test.inputHeaders) > 0 && test.expectedIP != "" {
					recorder := httptest.NewRecorder()
					request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost", nil)
					if err != nil {
						framework.Fatalf("error creating request: %s", err.Error())
					}

					for key, value := range test.inputHeaders {
						request.Header.Set(key, value)
					}

					trip.ServeHTTP(recorder, request)

					assertHeader(framework, request, "X-Real-Ip", test.expectedIP)
					assertHeader(framework, request, "X-Forwarded-For", test.expectedIP)
				}
			}
		})
	}
}

// assertHeader checks if the given header is present in the response and if it has the expected value.
func assertHeader(framework *testing.T, request *http.Request, header string, expected string) {
	framework.Helper()
	assert.Equal(framework, expected, request.Header.Get(header))
}
