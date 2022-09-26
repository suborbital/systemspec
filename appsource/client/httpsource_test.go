package client

import (
	"fmt"
	"github.com/suborbital/appspec/system"
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ system.Credential = (*PhonyCredential)(nil)

type PhonyCredential struct {
	scheme string
	value  string
}

func (p PhonyCredential) Scheme() string {
	return p.scheme
}

func (p PhonyCredential) Value() string {
	return p.value
}

func NewCredential(scheme, value string) system.Credential {
	return &PhonyCredential{
		scheme: scheme,
		value:  value,
	}
}

func TestAuthedRequest(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/system/v1/state":
			fmt.Printf("%+v\n", r.Header.Get(http.CanonicalHeaderKey("Authorization")))
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	}))

	defer svr.Close()

	source := NewHTTPSource(svr.URL, NewCredential("Bearer", "token"))

	source.State()
	//source.Overview()
	//source.TenantOverview("ident")
}
