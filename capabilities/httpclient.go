package capabilities

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

const defaultTimeout = 10 * time.Second

// HTTPConfig is configuration for the HTTP capability.
type HTTPConfig struct {
	Enabled bool      `json:"enabled" yaml:"enabled"`
	Rules   HTTPRules `json:"rules" yaml:"rules"`
}

// HTTPCapability gives Modules the ability to make HTTP requests.
type HTTPCapability interface {
	Do(auth AuthCapability, method, urlString string, body []byte, headers http.Header) (*http.Response, error)
}

type httpClient struct {
	config HTTPConfig
	client *http.Client
}

// DefaultHTTPClient creates an HTTP client with no restrictions.
func DefaultHTTPClient(config HTTPConfig) HTTPCapability {
	d := &httpClient{
		config: config,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	return d
}

// Do performs the provided request.
func (h *httpClient) Do(auth AuthCapability, method, urlString string, body []byte, headers http.Header) (*http.Response, error) {
	if !h.config.Enabled {
		return nil, ErrCapabilityNotEnabled
	}

	urlObj, err := url.Parse(urlString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to url.Parse")
	}

	ctx, cxl := context.WithTimeout(context.Background(), 5*time.Second)
	defer cxl()

	req, err := http.NewRequestWithContext(ctx, method, urlObj.String(), bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewRequest")
	}

	if err := h.config.Rules.requestIsAllowed(req); err != nil {
		return nil, errors.Wrap(err, "failed to requestIsAllowed")
	}

	authHeader := auth.HeaderForDomain(urlObj.Host)
	if authHeader != nil && authHeader.Value != "" {
		headers.Add("Authorization", fmt.Sprintf("%s %s", authHeader.HeaderType, authHeader.Value))
	}

	req.Header = headers

	return h.client.Do(req)
}
