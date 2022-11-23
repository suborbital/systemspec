package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/pkg/errors"

	fqmn "github.com/suborbital/systemspec/fqmn"
	"github.com/suborbital/systemspec/system"
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

// StateHandler is a handler to fetch the system State.
func (a *AppSourceVKRouter) StateHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		state, err := a.source.State()
		if err != nil {
			return vk.E(http.StatusInternalServerError, fmt.Sprintf("a.source.State(): %s", err.Error()))
		}

		return vk.RespondJSON(ctx.Context, w, state, http.StatusOK)
	}
}

// OverviewHandler is a handler to fetch the system overview.
func (a *AppSourceVKRouter) OverviewHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		overview, err := a.source.Overview()
		if err != nil {
			return vk.E(http.StatusInternalServerError, fmt.Sprintf("a.source.Overview(): %s", err.Error()))
		}

		return vk.RespondJSON(ctx.Context, w, overview, http.StatusOK)
	}
}

// TenantOverviewHandler is a handler to fetch a particular tenant's overview.
func (a *AppSourceVKRouter) TenantOverviewHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")

		tenantOverview, err := a.source.TenantOverview(ident)
		if err != nil {
			return vk.E(http.StatusInternalServerError, fmt.Sprintf("a.source.TenantOverview(%s): %s", ident, err.Error()))
		}

		return vk.RespondJSON(ctx.Context, w, tenantOverview, http.StatusOK)
	}
}

// GetModuleHandler is a handler to find a single module.
func (a *AppSourceVKRouter) GetModuleHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		ref := ctx.Params.ByName("ref")
		namespace := ctx.Params.ByName("namespace")
		mod := ctx.Params.ByName("mod")

		fqmnString, err := fqmn.FromParts(ident, namespace, mod, ref)
		if err != nil {
			ctx.Log.Error(errors.Wrap(err, "failed fqmn FromParts"))

			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		module, err := a.source.GetModule(fqmnString)
		if err != nil {
			ctx.Log.Error(errors.Wrap(err, "failed to GetFunction"))

			if errors.Is(err, system.ErrModuleNotFound) {
				return vk.Wrap(http.StatusNotFound, fmt.Errorf("failed to find function %s", fqmnString))
			} else if errors.Is(err, system.ErrAuthenticationFailed) {
				return vk.E(http.StatusUnauthorized, "unauthorized")
			}

			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, module, http.StatusOK)
	}
}

// WorkflowsHandler is a handler to fetch Workflows.
func (a *AppSourceVKRouter) WorkflowsHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		workflows, err := a.source.Workflows(ident, namespace, int64(version))
		if err != nil {
			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, workflows, http.StatusOK)
	}
}

// ConnectionsHandler is a handler to fetch Connection data.
func (a *AppSourceVKRouter) ConnectionsHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		connections, err := a.source.Connections(ident, namespace, int64(version))
		if err != nil {
			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, connections, http.StatusOK)
	}
}

// AuthenticationHandler is a handler to fetch Authentication data.
func (a *AppSourceVKRouter) AuthenticationHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		authentication, err := a.source.Authentication(ident, namespace, int64(version))
		if err != nil {
			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, authentication, http.StatusOK)
	}
}

// CapabilitiesHandler is a handler to fetch Capabilities data.
func (a *AppSourceVKRouter) CapabilitiesHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		caps, err := a.source.Capabilities(ident, namespace, int64(version))
		if err != nil {
			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, caps, http.StatusOK)
	}
}

// FileHandler is a handler to fetch Files.
func (a *AppSourceVKRouter) FileHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		filename := ctx.Params.ByName("filename")

		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		fileBytes, err := a.source.StaticFile(ident, int64(version), filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return vk.E(http.StatusNotFound, "not found")
			}

			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondBytes(ctx.Context, w, fileBytes, http.StatusOK)
	}
}

// QueriesHandler is a handler to fetch queries.
func (a *AppSourceVKRouter) QueriesHandler() vk.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, ctx *vk.Ctx) error {
		ident := ctx.Params.ByName("ident")
		namespace := ctx.Params.ByName("namespace")
		version, err := strconv.Atoi(ctx.Params.ByName("version"))
		if err != nil {
			return vk.E(http.StatusBadRequest, "bad request")
		}

		queries, err := a.source.Queries(ident, namespace, int64(version))
		if err != nil {
			return vk.E(http.StatusInternalServerError, "something went wrong")
		}

		return vk.RespondJSON(ctx.Context, w, queries, http.StatusOK)
	}
}
