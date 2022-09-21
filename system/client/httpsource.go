package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/fqmn"
	"github.com/suborbital/appspec/system"
	"github.com/suborbital/appspec/tenant"
)

// HTTPSource is an Source backed by an HTTP client connected to a remote source.
type HTTPSource struct {
	host       string
	authHeader string
	opts       system.Options
}

// NewHTTPSource creates a new HTTPSource that looks for a bundle at [host].
func NewHTTPSource(host string, creds system.CredentialSupplier) system.Source {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = fmt.Sprintf("http://%s", host)
	}

	h := &HTTPSource{
		host: host,
	}

	return h
}

// Start initializes the system source.
func (h *HTTPSource) Start(opts system.Options) error {
	h.opts = opts

	if err := h.pingServer(); err != nil {
		return errors.Wrap(err, "failed to pingServer")
	}

	return nil
}

// State returns the state of the entire system
func (h *HTTPSource) State() (*system.State, error) {
	s := &system.State{}
	if _, err := h.get("/system/v1/state", s); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /state"))
		return nil, errors.Wrap(err, "failed to get /state")
	}

	return s, nil
}

// Overview gets the overview for the entire system.
func (h *HTTPSource) Overview() (*system.Overview, error) {
	ovv := &system.Overview{}
	if _, err := h.get("/system/v1/overview", ovv); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /overview"))
		return nil, errors.Wrap(err, "failed to get /overview")
	}

	return ovv, nil
}

// TenantOverview gets the overview for a given tenant.
func (h *HTTPSource) TenantOverview(ident string) (*system.TenantOverview, error) {
	ovv := &system.TenantOverview{}

	if _, err := h.get(fmt.Sprintf("/system/v1/tenant/%s", ident), ovv); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get tenant overview"))
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
	if resp, err := h.authedGet(path, h.authHeader, module); err != nil {
		h.opts.Logger().Error(errors.Wrapf(err, "failed to get %s", path))

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, system.ErrAuthenticationFailed
		}

		return nil, system.ErrModuleNotFound
	}

	if h.authHeader != "" {
		// if we get this far, we assume the token was used to successfully get
		// the module from the control plane, and should therefore be used to
		// authenticate further calls for this function, so we cache its hash.
		module.TokenHash = system.TokenHash(h.authHeader)
	}

	return module, nil
}

// Workflows returns the Workflows for the system.
func (h *HTTPSource) Workflows(ident, namespace string, version int64) ([]tenant.Workflow, error) {
	workflows := make([]tenant.Workflow, 0)

	if _, err := h.get(fmt.Sprintf("/system/v1/workflows/%s/%s/%d", ident, namespace, version), &workflows); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /workflows"))
		return nil, errors.Wrap(err, "failed to get /schedules")
	}

	return workflows, nil
}

// Connections returns the Connections for the system.
func (h *HTTPSource) Connections(ident, namespace string, version int64) ([]tenant.Connection, error) {
	connections := []tenant.Connection{}

	if _, err := h.get(fmt.Sprintf("/system/v1/connections/%s/%s/%d", ident, namespace, version), &connections); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /connections"))
		return nil, errors.Wrap(err, "failed to get /connections")
	}

	return connections, nil
}

// Authentication returns the Authentication for the system.
func (h *HTTPSource) Authentication(ident, namespace string, version int64) (*tenant.Authentication, error) {
	authentication := &tenant.Authentication{}

	if _, err := h.get(fmt.Sprintf("/system/v1/authentication/%s/%s/%d", ident, namespace, version), authentication); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /authentication"))
	}

	return authentication, nil
}

// Capabilities returns the Capabilities for the system.
func (h *HTTPSource) Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error) {
	capabilities := &capabilities.CapabilityConfig{}

	if _, err := h.get(fmt.Sprintf("/system/v1/capabilities/%s/%s/%d", ident, namespace, version), capabilities); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /capabilities"))
		return nil, errors.Wrap(err, "failed to get /capabilities")
	}

	return capabilities, nil
}

// StaticFile returns a requested file.
func (h *HTTPSource) StaticFile(ident string, version int64, filename string) ([]byte, error) {
	path := fmt.Sprintf("/system/v1/file/%s/%d/%s", ident, version, filename)

	resp, err := h.get(path, nil)
	if err != nil {
		h.opts.Logger().Error(errors.Wrapf(err, "failed to get %s", path))
		return nil, os.ErrNotExist
	}

	defer resp.Body.Close()
	file, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to ReadAll")
	}

	return file, nil
}

// Queries returns the Queries for the system.
func (h *HTTPSource) Queries(ident, namespace string, version int64) ([]tenant.DBQuery, error) {
	queries := make([]tenant.DBQuery, 0)

	if _, err := h.get(fmt.Sprintf("/system/v1/queries/%s/%s/%d", ident, namespace, version), &queries); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /queries"))
		return nil, errors.Wrap(err, "failed to get /queries")
	}

	return queries, nil
}

// pingServer loops forever until it finds a server at the configured host.
func (h *HTTPSource) pingServer() error {
	for {
		if _, err := h.get("/system/v1/state", nil); err != nil {

			h.opts.Logger().Warn("failed to connect to remote source, will retry:", err.Error())

			time.Sleep(time.Second)

			continue
		}

		h.opts.Logger().Info("connected to remote source at", h.host)

		break
	}

	return nil
}

// get performs a GET request against the configured host and given path.
func (h *HTTPSource) get(path string, dest interface{}) (*http.Response, error) {
	return h.authedGet(path, "", dest)
}

// authedGet performs a GET request against the configured host and given path with the given auth header.
func (h *HTTPSource) authedGet(path, auth string, dest interface{}) (*http.Response, error) {
	url, err := url.Parse(fmt.Sprintf("%s%s", h.host, path))
	if err != nil {
		return nil, errors.Wrap(err, "failed to url.Parse")
	}

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to NewRequest")
	}

	if auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to Do request")
	}

	if resp.StatusCode != http.StatusOK {
		return resp, fmt.Errorf("response returned non-200 status: %d", resp.StatusCode)
	}

	if dest != nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed to ReadAll body")
		}

		if err := json.Unmarshal(body, dest); err != nil {
			return nil, errors.Wrap(err, "failed to json.Unmarshal")
		}
	}

	return resp, nil
}
