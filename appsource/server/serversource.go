package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/suborbital/appspec/appsource"
	fqmn "github.com/suborbital/appspec/fqmn"
	"github.com/suborbital/vektor/vk"
)

// AppSourceVKRouter is a helper struct to generate a VK router that can serve
// an HTTP AppSource based on an actual AppSource object.
type AppSourceVKRouter struct {
	appSource appsource.AppSource
	options   appsource.Options
}

// NewAppSourceVKRouter creates a new AppSourceVKRouter.
func NewAppSourceVKRouter(appSource appsource.AppSource, opts appsource.Options) *AppSourceVKRouter {
	h := &AppSourceVKRouter{
		appSource: appSource,
		options:   opts,
	}

	return h
}

// GenerateRouter generates a VK router that uses an AppSource to serve data.
func (a *AppSourceVKRouter) GenerateRouter() (*vk.Router, error) {
	if err := a.appSource.Start(a.options); err != nil {
		return nil, errors.Wrap(err, "failed to appSource.Start")
	}

	router := vk.NewRouter(a.options.Logger(), "")

	v1 := vk.Group("/appsource/v1")

	v1.GET("/state", a.StateHandler())
	v1.GET("/overview", a.OverviewHandler())
	v1.GET("/tenant/:ident", a.TenantOverviewHandler())
	v1.GET("/module/:ident/:ref/:namespace/:mod", a.GetModuleHandler())
	v1.GET("/workflows/:ident/:namespace/:version", a.WorkflowsHandler())
	v1.GET("/connections/:ident/:namespace/:version", a.ConnectionsHandler())
	v1.GET("/authentication/:ident/:namespace/:version", a.AuthenticationHandler())
	v1.GET("/capabilities/:ident/:namespace/:version", a.CapabilitiesHandler())
	v1.GET("/queries/:ident/:namespace/:version", a.QueriesHandler())

	v1.GET("/file/:ident/:version/*filename", a.FileHandler())

	router.AddGroup(v1)

	return router, nil
}

// State is a handler to fetch the system State.
func (a *AppSourceVKRouter) StateHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		return a.appSource.State()
	}
}

// OverviewHandler is a handler to fetch the system overview.
func (a *AppSourceVKRouter) OverviewHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		return a.appSource.Overview()
	}
}

// TenantOverviewHandler is a handler to fetch a particular tenant's overview.
func (a *AppSourceVKRouter) TenantOverviewHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")

		return a.appSource.TenantOverview(ident)
	}
}

// GetModuleHandler is a handler to find a single module.
func (a *AppSourceVKRouter) GetModuleHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		ref := ctx.Params.ByName("ref")
		namespace := ctx.Params.ByName("namespace")
		mod := ctx.Params.ByName("mod")

		fqmnString, err := fqmn.FromParts(ident, namespace, mod, ref)
		if err != nil {
			ctx.Log.Error(errors.Wrap(err, "failed fqmn FromParts"))

			return nil, vk.E(http.StatusInternalServerError, "something went wrong")
		}

		runnable, err := a.appSource.GetModule(fqmnString)
		if err != nil {
			ctx.Log.Error(errors.Wrap(err, "failed to GetFunction"))

			if errors.Is(err, appsource.ErrModuleNotFound) {
				return nil, vk.Wrap(http.StatusNotFound, fmt.Errorf("failed to find function %s", fqmnString))
			} else if errors.Is(err, appsource.ErrAuthenticationFailed) {
				return nil, vk.E(http.StatusUnauthorized, "unauthorized")
			}

			return nil, vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return runnable, nil
	}
}

// WorkflowsHandler is a handler to fetch Workflows.
func (a *AppSourceVKRouter) WorkflowsHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}

		return a.appSource.Workflows(ident, namespace, int64(version))
	}
}

// ConnectionsHandler is a handler to fetch Connection data.
func (a *AppSourceVKRouter) ConnectionsHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}

		return a.appSource.Connections(ident, namespace, int64(version))
	}
}

// AuthenticationHandler is a handler to fetch Authentication data.
func (a *AppSourceVKRouter) AuthenticationHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}

		return a.appSource.Authentication(ident, namespace, int64(version))
	}
}

// CapabilitiesHandler is a handler to fetch Capabilities data.
func (a *AppSourceVKRouter) CapabilitiesHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}

		return a.appSource.Capabilities(ident, namespace, int64(version))
	}
}

// FileHandler is a handler to fetch Files.
func (a *AppSourceVKRouter) FileHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		filename := ctx.Params.ByName("filename")

		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}

		fileBytes, err := a.appSource.StaticFile(ident, int64(version), filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, vk.E(http.StatusNotFound, "not found")
			}

			return nil, vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return fileBytes, nil
	}
}

// QueriesHandler is a handler to fetch queries.
func (a *AppSourceVKRouter) QueriesHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return nil, vk.E(http.StatusBadRequest, "bad request")
		}
		return a.appSource.Queries(ident, namespace, int64(version))
	}
}
