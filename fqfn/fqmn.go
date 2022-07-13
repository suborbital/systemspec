package fqmn

import (
	"errors"
	"fmt"
	"strings"
)

////////////////////////////////////////////////////////////////////////////////////
// An FQMN (fully-qualified function name) is a "globally unique"
// name for a specific function from a specific application ref
// example: fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api-users/add-user
// i.e. fqmn://<identifier>/<ref>/<namespace>/<funcname>
//
// These URI forms are also supported:
//
// 		/api/users/add-user
// 		(single-application function)
//
// 		/ref/f0e4c2f76c58916ec258f246851be/api/users/add-user
// 		(reference to a version of a single-application function)
//
// 		/com.suborbital.acmeco/api/users/add-user
// 		(multi-application, single-domain function)
//
// 		/ref/f0e4c2f76c58916ec258f246851be/com.suborbital.acmeco/api/users/add-user
// 		(reference to a version of a multi-application, single-domain function)
//
// Additionally, a URL form assumes the function identifier is the reverse domain of
// the URL, but otherwise is the same as the URI form.
// example: https://acmeco.suborbital.com/api-users/add-user
////////////////////////////////////////////////////////////////////////////////////

// NamespaceDefault and others represent conts for namespaces.
const (
	NamespaceDefault = "default"
)

// FQMN is a parsed fqmn.
type FQMN struct {
	Identifier string `json:"identifier"`
	Namespace  string `json:"namespace"`
	Fn         string `json:"fn"`
	Ref        string `json:"ref"`
}

var errWrongPrefix = errors.New("FQMN must begin with 'fqmn://' or '/'")
var errMustBeFullyQualified = errors.New("FQMN text format must contain an identifier, ref, namespace, and function name")
var errTooFewParts = errors.New("FQMN must contain at least a namespace and function name")
var errTrailingSlash = errors.New("FQMN must not end in a trailing slash")

func Parse(fqmnString string) (FQMN, error) {
	if strings.HasPrefix(fqmnString, "fqmn://") {
		return parseTextFormat(fqmnString)
	}

	if strings.HasPrefix(fqmnString, "/") {
		return parseUriFormat(fqmnString)
	}

	return FQMN{}, errWrongPrefix
}

func parseTextFormat(fqmnString string) (FQMN, error) {
	fqmnString = strings.TrimPrefix(fqmnString, "fqmn://")

	segments := strings.Split(fqmnString, "/")

	// There should be at least four segments representing the ident, ref, namespace, and name.
	// Additional segments would be the result of multi-level namespaces.
	if len(segments) < 4 {
		return FQMN{}, errMustBeFullyQualified
	}

	// If the last segment is empty, there was a trailing slash
	if segments[len(segments)-1] == "" {
		return FQMN{}, errTrailingSlash
	}

	identifier := segments[0]

	ref := segments[1]

	// Reconstruct the namespace
	namespace := strings.Join(segments[2:len(segments)-1], "/")

	// The function name is just the last element
	fn := segments[len(segments)-1]

	fqmn := FQMN{
		Identifier: identifier,
		Namespace:  namespace,
		Fn:         fn,
		Ref:        ref,
	}

	return fqmn, nil
}

func parseUriFormat(fqmnString string) (FQMN, error) {
	segments := strings.Split(fqmnString, "/")
	// The first segment will be empty since the string starts with '/'
	segments = segments[1:]

	// There should be at least two segments
	if len(segments) < 2 {
		return FQMN{}, errTooFewParts
	}

	// If the last segment is empty, there was a trailing slash
	if segments[len(segments)-1] == "" {
		return FQMN{}, errTrailingSlash
	}

	// Check for a ref
	var ref string
	if segments[0] == "ref" {
		ref = segments[1]
		segments = segments[2:]

		// There should be at least two more segments
		if len(segments) < 2 {
			return FQMN{}, errTooFewParts
		}
	}

	// Check for an identifier
	var identifier string
	if strings.Count(segments[0], ".") == 2 {
		identifier = segments[0]
		segments = segments[1:]

		// There _still_ should be at least two more segments
		if len(segments) < 2 {
			return FQMN{}, errTooFewParts
		}
	}

	// Reconstruct the namespace
	namespace := strings.Join(segments[:len(segments)-1], "/")

	// The function name is just the last element
	fn := segments[len(segments)-1]

	fqmn := FQMN{
		Identifier: identifier,
		Namespace:  namespace,
		Fn:         fn,
		Ref:        ref,
	}

	return fqmn, nil
}

func MigrateV1ToV2(name, ref string) (FQMN, error) {
	// Parse V1 format and swap version for ref

	// if the name contains a #, parse that out as the identifier.
	identifier := ""
	identParts := strings.SplitN(name, "#", 2)
	if len(identParts) == 2 {
		identifier = identParts[0]
		name = identParts[1]
	}

	// if a Runnable is referenced with its namespace, i.e. users#getUser
	// then we need to parse that and ensure we only match that namespace.

	namespace := NamespaceDefault
	namespaceParts := strings.SplitN(name, "::", 2)
	if len(namespaceParts) == 2 {
		namespace = namespaceParts[0]
		name = namespaceParts[1]
	}

	// next, if the name contains an @, parse the name and ignore app version.
	versionParts := strings.SplitN(name, "@", 2)
	if len(versionParts) == 2 {
		name = versionParts[0]
	}

	fqmn := FQMN{
		Identifier: identifier,
		Namespace:  namespace,
		Fn:         name,
		Ref:        ref,
	}

	return fqmn, nil
}

// HeadlessURLPath returns the headless URL path for a function.
func (f FQMN) HeadlessURLPath() string {
	return fmt.Sprintf("/%s/%s/%s/%s", f.Identifier, f.Namespace, f.Fn, f.Ref)
}

func FromParts(ident, namespace, fn, ref string) string {
	return fmt.Sprintf("fqmn://%s/%s/%s/%s", ident, ref, namespace, fn)
}
