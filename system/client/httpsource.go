package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/suborbital/systemspec/capabilities"
	"github.com/suborbital/systemspec/fqmn"
	"github.com/suborbital/systemspec/system"
	"github.com/suborbital/systemspec/tenant"
)

const defaultTimeout = 10 * time.Second

// HTTPSource is a Source backed by an HTTP client connected to a remote source.
type HTTPSource struct {
	host       string
	authHeader string
	client     *http.Client
}

// NewHTTPSource creates a new HTTPSource that looks for a bundle at [host].
func NewHTTPSource(hostIn string, creds system.Credential) system.Source {
	host := hostIn
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = fmt.Sprintf("http://%s", host)
	}

	source := &HTTPSource{
		host: host,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	if creds != nil {
		source.authHeader = fmt.Sprintf("%s %s", creds.Scheme(), creds.Value())
	}

	return source
}

// Start initializes the system source.
func (h *HTTPSource) Start() error {
	if err := h.pingServer(); err != nil {
		return errors.Wrap(err, "failed to pingServer")
	}

	return nil
}

// State returns the state of the entire system.
func (h *HTTPSource) State() (*system.State, error) {
	s := &system.State{}
	if err := h.get("/system/v1/state", s); err != nil {
		return nil, errors.Wrap(err, "failed to get /state")
	}

	return s, nil
}

// Overview gets the overview for the entire system.
func (h *HTTPSource) Overview() (*system.Overview, error) {
	ovv := &system.Overview{}
	if err := h.get("/system/v1/overview", ovv); err != nil {
		return nil, errors.Wrap(err, "failed to get /overview")
	}

	return ovv, nil
}

// TenantOverview gets the overview for a given tenant.
func (h *HTTPSource) TenantOverview(ident string) (*system.TenantOverview, error) {
	ovv := &system.TenantOverview{}

	if err := h.get(fmt.Sprintf("/system/v1/tenant/%s", ident), ovv); err != nil {
		return nil, errors.Wrap(err, "failed to get tenant overview")
	}

	return ovv, nil
}

// GetModule returns a nil error if a Module with the
// provided FQMN can be made available at the next sync,
// otherwise ErrRunnableNotFound is returned.
func (h *HTTPSource) GetModule(FQMN string) (*tenant.Module, error) {
	f, err := fqmn.Parse(FQMN)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Parse FQMN")
	}

	path := fmt.Sprintf("/system/v1/module%s", f.URLPath())

	module := &tenant.Module{}
	if err := h.authedGet(path, h.authHeader, module); err != nil {
		if errors.Is(err, system.ErrAuthenticationFailed) {
			return nil, errors.Wrap(err, system.ErrAuthenticationFailed.Error())
		}

		return nil, errors.Wrap(err, system.ErrModuleNotFound.Error())
	}

	return module, nil
}

// Workflows returns the Workflows for the system.
func (h *HTTPSource) Workflows(ident, namespace string, version int64) ([]tenant.Workflow, error) {
	workflows := make([]tenant.Workflow, 0)

	if err := h.get(fmt.Sprintf("/system/v1/workflows/%s/%s/%d", ident, namespace, version), &workflows); err != nil {
		return nil, errors.Wrap(err, "failed to get /schedules")
	}

	return workflows, nil
}

// Connections returns the Connections for the system.
func (h *HTTPSource) Connections(ident, namespace string, version int64) ([]tenant.Connection, error) {
	connections := make([]tenant.Connection, 0)

	if err := h.get(fmt.Sprintf("/system/v1/connections/%s/%s/%d", ident, namespace, version), &connections); err != nil {
		return nil, errors.Wrap(err, "failed to get /connections")
	}

	return connections, nil
}

// Authentication returns the Authentication for the system.
func (h *HTTPSource) Authentication(ident, namespace string, version int64) (*tenant.Authentication, error) {
	authentication := &tenant.Authentication{}

	if err := h.get(fmt.Sprintf("/system/v1/authentication/%s/%s/%d", ident, namespace, version), authentication); err != nil {
		return nil, errors.Wrap(err, "failed to get /authentication")
	}

	return authentication, nil
}

// Capabilities returns the Capabilities for the system.
func (h *HTTPSource) Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error) {
	caps := &capabilities.CapabilityConfig{}

	if err := h.get(fmt.Sprintf("/system/v1/caps/%s/%s/%d", ident, namespace, version), caps); err != nil {
		return nil, errors.Wrap(err, "failed to get /caps")
	}

	return caps, nil
}

// pingServer loops forever until it finds a server at the configured host.
func (h *HTTPSource) pingServer() error {
	for {
		if err := h.get("/system/v1/state", nil); err != nil {
			time.Sleep(time.Second)

			continue
		}

		break
	}

	return nil
}

// get performs a GET request against the configured host and given path.
func (h *HTTPSource) get(path string, dest any) error {
	return h.authedGet(path, h.authHeader, dest)
}

// authedGet performs a GET request against the configured host and given path with the given auth header.
func (h *HTTPSource) authedGet(path, auth string, dest any) error {
	parsedURL, err := url.Parse(fmt.Sprintf("%s%s", h.host, path))
	if err != nil {
		return errors.Wrap(err, "failed to parsedURL.Parse")
	}

	ctx, cxl := context.WithTimeout(context.Background(), defaultTimeout)
	defer cxl()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return errors.Wrap(err, "failed to NewRequest")
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to Do request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to ReadAll body")
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return errors.WithMessage(system.ErrAuthenticationFailed, fmt.Sprintf("response body: %s", string(body)))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response returned non-200 status: %d with error message: %s", resp.StatusCode, string(body))
	}

	if dest != nil {
		if err := json.Unmarshal(body, dest); err != nil {
			return errors.Wrap(err, "failed to json.Unmarshal")
		}
	}

	return nil
}
