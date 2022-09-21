package fqmn

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

////////////////////////////////////////////////////////////////////////////////////
// An FQMN (fully-qualified module name) is a "globally unique"
// name for a specific module from a specific tenant ref
// example: fqmn://suborbital.acmeco/api-users/add-user@98qhrfgo3089hafrouhqf48
// i.e. fqmn://<tenant>/<namespace>/<modname>@<ref>
//
// Namespaces can be deeply nested.
//
// These URI forms are also supported:
//
//      /name/<namespace>/<modname>
// 		e.g. /name/api/users/add-user
// 		(addressing a module by name)
//
// 		/ref/f0e4c2f76c58916ec258f246851be
//      e.g. /ref/<ref>
// 		(addressing a module by module hash)
//
////////////////////////////////////////////////////////////////////////////////////

// NamespaceDefault and others represent conts for namespaces.
const (
	NamespaceDefault = "default"
)

// FQMN is a parsed fqmn.
type FQMN struct {
	Tenant    string `json:"tenant"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ref       string `json:"ref"`
}

var ErrFQMNParseFailure = errors.New("FQMN failed to parse")
var ErrFQMNConstructionFailure = errors.New("All FQMN must be defined")

var errWrongPrefix = errors.Wrap(ErrFQMNParseFailure, "FQMN must begin with 'fqmn://', '/name', or '/ref'")
var errMustBeFullyQualified = errors.Wrap(ErrFQMNParseFailure, "FQMN text format must contain an tenant, ref, namespace, and module name")
var errTooFewParts = errors.Wrap(ErrFQMNParseFailure, "FQMN must contain a namespace and module name")
var errMalformedRef = errors.Wrap(ErrFQMNParseFailure, "'/ref' format may only contain one reference")
var errTrailingSlash = errors.Wrap(ErrFQMNParseFailure, "FQMN must not end in a trailing slash")

func Parse(fqmnString string) (FQMN, error) {
	if strings.HasPrefix(fqmnString, "fqmn://") {
		return parseTextFormat(fqmnString)
	}

	if strings.HasPrefix(fqmnString, "/name/") {
		return parseNameUri(fqmnString)
	}

	if strings.HasPrefix(fqmnString, "/ref/") {
		return parseRefUri(fqmnString)
	}

	return FQMN{}, errors.Wrapf(errWrongPrefix, "failed to parse string %q", fqmnString)
}

func parseTextFormat(fqmnString string) (FQMN, error) {
	fqmnString = strings.TrimPrefix(fqmnString, "fqmn://")

	refSegments := strings.SplitN(fqmnString, "@", 2)
	var ref string
	if len(refSegments) == 2 {
		ref = refSegments[1]
	}

	fqmnString = refSegments[0]

	segments := strings.Split(fqmnString, "/")

	// There should be at least three segments representing the tenant, namespace, and modname.
	// Additional segments would be the result of multi-level namespaces.
	if len(segments) < 3 {
		return FQMN{}, errMustBeFullyQualified
	}

	// If the last segment is empty, there was a trailing slash
	if segments[len(segments)-1] == "" {
		return FQMN{}, errTrailingSlash
	}

	tenant := segments[0]

	// Reconstruct the namespace
	namespace := strings.Join(segments[1:len(segments)-1], "/")

	// The module name is just the last element
	name := segments[len(segments)-1]

	fqmn := FQMN{
		Tenant:    tenant,
		Namespace: namespace,
		Name:      name,
		Ref:       ref,
	}

	return fqmn, nil
}

func parseNameUri(fqmnString string) (FQMN, error) {
	fqmnString = strings.TrimPrefix(fqmnString, "/name/")
	segments := strings.Split(fqmnString, "/")

	// There should be at least two segments
	if len(segments) < 2 {
		return FQMN{}, errTooFewParts
	}

	// If the last segment is empty, there was a trailing slash
	if segments[len(segments)-1] == "" {
		return FQMN{}, errTrailingSlash
	}

	// Reconstruct the namespace
	namespace := strings.Join(segments[:len(segments)-1], "/")

	// The function name is just the last element
	name := segments[len(segments)-1]

	fqmn := FQMN{
		Namespace: namespace,
		Name:      name,
	}

	return fqmn, nil
}

func parseRefUri(fqmnString string) (FQMN, error) {
	fqmnString = strings.TrimPrefix(fqmnString, "/ref/")
	segments := strings.Split(fqmnString, "/")

	// If the last segment is empty, there was a trailing slash
	if segments[len(segments)-1] == "" {
		return FQMN{}, errTrailingSlash
	}

	// There should be only one segment
	if len(segments) != 1 {
		return FQMN{}, errMalformedRef
	}

	ref := segments[0]

	fqmn := FQMN{
		Ref: ref,
	}

	return fqmn, nil
}

func MigrateV1ToV2(name, ref string) (FQMN, error) {
	// Parse V1 format and swap version for ref

	// if the name contains a #, parse that out as the tenant.
	tenant := ""
	tenantParts := strings.SplitN(name, "#", 2)
	if len(tenantParts) == 2 {
		tenant = tenantParts[0]
		name = tenantParts[1]
	}

	// if a Module is referenced with its namespace, i.e. users#getUser
	// then we need to parse that and ensure we only match that namespace.

	namespace := NamespaceDefault
	namespaceParts := strings.SplitN(name, "::", 2)
	if len(namespaceParts) == 2 {
		namespace = namespaceParts[0]
		name = namespaceParts[1]
	}

	// next, if the name contains an @, parse the name and ignore tenant version.
	versionParts := strings.SplitN(name, "@", 2)
	if len(versionParts) == 2 {
		name = versionParts[0]
	}

	fqmn := FQMN{
		Tenant:    tenant,
		Namespace: namespace,
		Name:      name,
		Ref:       ref,
	}

	return fqmn, nil
}

// URLPath returns the URL path for a function.
func (f FQMN) URLPath() string {
	return fmt.Sprintf("/%s/%s/%s/%s", f.Tenant, f.Ref, f.Namespace, f.Name)
}

// FromParts returns an FQMN from the provided parts
func FromParts(tenant, namespace, module, ref string) (string, error) {
	if tenant == "" || namespace == "" || module == "" || ref == "" {
		return "", ErrFQMNConstructionFailure
	}
	return fmt.Sprintf("fqmn://%s/%s/%s@%s", tenant, namespace, module, ref), nil
}
