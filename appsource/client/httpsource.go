package appsource

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

	"github.com/suborbital/appspec/appsource"
	"github.com/suborbital/appspec/capabilities"
	"github.com/suborbital/appspec/directive"
	"github.com/suborbital/appspec/fqfn"
)

// HTTPSource is an AppSource backed by an HTTP client connected to a remote source.
type HTTPSource struct {
	host       string
	authHeader string
	opts       appsource.Options
}

// NewHTTPSource creates a new HTTPSource that looks for a bundle at [host].
func NewHTTPSource(host string, creds appsource.CredentialSupplier) appsource.AppSource {
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = fmt.Sprintf("http://%s", host)
	}

	h := &HTTPSource{
		host: host,
	}

	return h
}

// Start initializes the app source.
func (h *HTTPSource) Start(opts appsource.Options) error {
	h.opts = opts

	if err := h.pingServer(); err != nil {
		return errors.Wrap(err, "failed to pingServer")
	}

	return nil
}

// State returns the state of the entire system
func (h *HTTPSource) State() (*appsource.State, error) {
	s := &appsource.State{}
	if _, err := h.get("/state", s); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /state"))
		return nil, errors.Wrap(err, "failed to get /state")
	}

	return s, nil
}

// Overview gets the overview for the entire system.
func (h *HTTPSource) Overview() (*appsource.Overview, error) {
	ovv := &appsource.Overview{}
	if _, err := h.get("/overview", ovv); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /overview"))
		return nil, errors.Wrap(err, "failed to get /overview")
	}

	return ovv, nil
}

// TenantOverview gets the overview for a given tenant.
func (h *HTTPSource) TenantOverview(ident string) (*appsource.TenantOverview, error) {
	ovv := &appsource.TenantOverview{}

	if _, err := h.get(fmt.Sprintf("/tenant/%s", ident), ovv); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get tenant overview"))
		return nil, errors.Wrap(err, "failed to get tenant overview")
	}

	return ovv, nil
}

// GetModule returns a nil error if a Runnable with the
// provided FQFN can be made available at the next sync,
// otherwise ErrRunnableNotFound is returned.
func (h *HTTPSource) GetModule(FQFN string) (*appsource.Module, error) {
	f := fqfn.Parse(FQFN)

	path := fmt.Sprintf("/module%s", f.HeadlessURLPath())

	runnable := directive.Runnable{}
	if resp, err := h.authedGet(path, h.authHeader, &runnable); err != nil {
		h.opts.Logger().Error(errors.Wrapf(err, "failed to get %s", path))

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, appsource.ErrAuthenticationFailed
		}

		return nil, appsource.ErrModuleNotFound
	}

	if h.authHeader != "" {
		// if we get this far, we assume the token was used to successfully get
		// the runnable from the control plane, and should therefore be used to
		// authenticate further calls for this function, so we cache its hash.
		runnable.TokenHash = appsource.TokenHash(h.authHeader)
	}

	m := &appsource.Module{
		Name:      runnable.Name,
		Namespace: runnable.Namespace,
		Ref:       "",
		FQFN:      runnable.FQFN,
		Revisions: []appsource.ModuleRevision{},
	}

	return m, nil
}

// Workflows returns the Workflows for the app.
func (h *HTTPSource) Workflows(ident, namespace string, version int64) ([]directive.Schedule, error) {
	workflows := make([]directive.Schedule, 0)

	if _, err := h.get(fmt.Sprintf("/workflows/%s/%s/%d", ident, namespace, version), &workflows); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /workflows"))
		return nil, errors.Wrap(err, "failed to get /schedules")
	}

	return workflows, nil
}

// Connections returns the Connections for the app.
func (h *HTTPSource) Connections(ident, namespace string, version int64) (*directive.Connections, error) {
	connections := &directive.Connections{}

	if _, err := h.get(fmt.Sprintf("/connections/%s/%s/%d", ident, namespace, version), connections); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /connections"))
		return nil, errors.Wrap(err, "failed to get /connections")
	}

	return connections, nil
}

// Authentication returns the Authentication for the app.
func (h *HTTPSource) Authentication(ident, namespace string, version int64) (*directive.Authentication, error) {
	authentication := &directive.Authentication{}

	if _, err := h.get(fmt.Sprintf("/authentication/%s/%s/%d", ident, namespace, version), authentication); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /authentication"))
	}

	return authentication, nil
}

// Capabilities returns the Capabilities for the app.
func (h *HTTPSource) Capabilities(ident, namespace string, version int64) (*capabilities.CapabilityConfig, error) {
	capabilities := &capabilities.CapabilityConfig{}

	if _, err := h.get(fmt.Sprintf("/capabilities/%s/%s/%d", ident, namespace, version), capabilities); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /capabilities"))
		return nil, errors.Wrap(err, "failed to get /capabilities")
	}

	return capabilities, nil
}

// StaticFile returns a requested file.
func (h *HTTPSource) StaticFile(ident, namespace, filename string, version int64) ([]byte, error) {
	path := fmt.Sprintf("/file/%s/%s/%s/%d", ident, namespace, filename, version)

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

// Queries returns the Queries for the app.
func (h *HTTPSource) Queries(ident, namespace string, version int64) ([]directive.DBQuery, error) {
	queries := make([]directive.DBQuery, 0)

	if _, err := h.get(fmt.Sprintf("/queries/%s/%s/%d", ident, namespace, version), &queries); err != nil {
		h.opts.Logger().Error(errors.Wrap(err, "failed to get /queries"))
		return nil, errors.Wrap(err, "failed to get /queries")
	}

	return queries, nil
}

// pingServer loops forever until it finds a server at the configured host.
func (h *HTTPSource) pingServer() error {
	for {
		if _, err := h.get("/meta", nil); err != nil {

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
