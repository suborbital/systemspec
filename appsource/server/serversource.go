package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"

	fqmn "github.com/suborbital/appspec/fqmn"
	"github.com/suborbital/appspec/system"
	"github.com/suborbital/vektor/vk"
)

// AppSourceVKRouter is a helper struct to generate a VK router that can serve
// an HTTP Source based on an actual Source object.
type AppSourceVKRouter struct {
	source  system.Source
	options system.Options
}

// NewAppSourceVKRouter creates a new AppSourceVKRouter.
func NewAppSourceVKRouter(source system.Source, opts system.Options) *AppSourceVKRouter {
	h := &AppSourceVKRouter{
		source:  source,
		options: opts,
	}

	return h
}

// GenerateRouter generates a VK router that uses a Source to serve data.
func (a *AppSourceVKRouter) GenerateRouter() (*vk.Router, error) {
	if err := a.source.Start(a.options); err != nil {
		return nil, errors.Wrap(err, "failed to source.Start")
	}

	router := vk.NewRouter(a.options.Logger(), "")

	v1 := vk.Group("/system/v1")

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
		return a.source.State()
	}
}

// OverviewHandler is a handler to fetch the system overview.
func (a *AppSourceVKRouter) OverviewHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		return a.source.Overview()
	}
}

// TenantOverviewHandler is a handler to fetch a particular tenant's overview.
func (a *AppSourceVKRouter) TenantOverviewHandler() vk.HandlerFunc {
	return func(r *http.Request, ctx *vk.Ctx) (interface{}, error) {
		ident := ctx.Params.ByName("ident")

		return a.source.TenantOverview(ident)
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

		module, err := a.source.GetModule(fqmnString)
		if err != nil {
			ctx.Log.Error(errors.Wrap(err, "failed to GetFunction"))

			if errors.Is(err, system.ErrModuleNotFound) {
				return nil, vk.Wrap(http.StatusNotFound, fmt.Errorf("failed to find function %s", fqmnString))
			} else if errors.Is(err, system.ErrAuthenticationFailed) {
				return nil, vk.E(http.StatusUnauthorized, "unauthorized")
			}

			return nil, vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return module, nil
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

		return a.source.Workflows(ident, namespace, int64(version))
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

		return a.source.Connections(ident, namespace, int64(version))
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

		return a.source.Authentication(ident, namespace, int64(version))
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

		return a.source.Capabilities(ident, namespace, int64(version))
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

		fileBytes, err := a.source.StaticFile(ident, int64(version), filename)
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
		return a.source.Queries(ident, namespace, int64(version))
	}
}
