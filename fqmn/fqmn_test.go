package fqmn

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FQMNSuite struct {
	suite.Suite
}

func TestFQMNSuite(t *testing.T) {
	suite.Run(t, &FQMNSuite{})
}

func (s *FQMNSuite) TestParse() {
	for _, tt := range []struct {
		name string
		text string
		fqmn FQMN
		error
	}{
		{"text format single-level namespace", "fqmn://suborbital.acmeco/api-users/add-user", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api-users",
			Name:      "add-user",
		}, nil},
		{"text format two-level namespace", "fqmn://suborbital.acmeco/api/users/add-user", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api/users",
			Name:      "add-user",
		}, nil},
		{"text format multi-level namespace", "fqmn://suborbital.acmeco/api/users/auora/add-user", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api/users/auora",
			Name:      "add-user",
		}, nil},
		{"text format single-level namespace with ref", "fqmn://suborbital.acmeco/api-users/add-user@98qhrfgo3089hafrouhqf48", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api-users",
			Name:      "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"text format two-level namespace with ref", "fqmn://suborbital.acmeco/api/users/add-user@98qhrfgo3089hafrouhqf48", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api/users",
			Name:      "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"text format multi-level namespace with ref", "fqmn://suborbital.acmeco/api/users/auora/add-user@98qhrfgo3089hafrouhqf48", FQMN{
			Tenant:    "suborbital.acmeco",
			Namespace: "api/users/auora",
			Name:      "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"module by name with single-level namespace", "/name/api-users/add-user", FQMN{
			Namespace: "api-users",
			Name:      "add-user",
		}, nil},
		{"module by name with two-level namespace", "/name/api/users/add-user", FQMN{
			Namespace: "api/users",
			Name:      "add-user",
		}, nil},
		{"module by name with multi-level namespace", "/name/api/users/auora/add-user", FQMN{
			Namespace: "api/users/auora",
			Name:      "add-user",
		}, nil},
		{"by reference", "/ref/98qhrfgo3089hafrouhqf48", FQMN{
			Ref: "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"malformed—doesn't start with right prefix 1", "fqmn:suborbital.acmeco/api/users/auora/add-user@98qhrfgo3089hafrouhqf48", FQMN{}, errWrongPrefix},
		{"malformed—doesn't start with right prefix 2", "https://suborbital.acmeco/api/users/auora/add-user@98qhrfgo3089hafrouhqf48", FQMN{}, errWrongPrefix},
		{"malformed—doesn't start with right prefix 3", "/module/api/users/auora/add-user", FQMN{}, errWrongPrefix},
		{"malformed—doesn't start with right prefix 4", "/reference/98qhrfgo3089hafrouhqf48", FQMN{}, errWrongPrefix},
		{"malformed—not fully-qualified 1", "fqmn://suborbital.acmeco@98qhrfgo3089hafrouhqf48", FQMN{}, errMustBeFullyQualified},
		{"malformed—not fully-qualified 2", "fqmn://suborbital.acmeco/add-user@98qhrfgo3089hafrouhqf48", FQMN{}, errMustBeFullyQualified},
		{"malformed—not enough parts 1", "/name/add-user", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 2", "/name/suborbital.acmeco", FQMN{}, errTooFewParts},
		{"malformed—not a reference", "/ref/98qhrfgo3089hafrouhqf48/add-user", FQMN{}, errMalformedRef},
		{"malformed—not a reference", "/ref/98qhrfgo3089hafrouhqf48/suborbital.acmeco", FQMN{}, errMalformedRef},
		{"malformed—not a reference", "/ref/98qhrfgo3089hafrouhqf48/suborbital.acmeco/add-user", FQMN{}, errMalformedRef},
		{"malformed—trailing slash 1", "fqmn://suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/add-user/", FQMN{}, errTrailingSlash},
		{"malformed—trailing slash 2", "/ref/98qhrfgo3089hafrouhqf48/", FQMN{}, errTrailingSlash},
	} {
		s.Run(tt.name, func() {
			fqmn, err := Parse(tt.text)

			if err != nil {
				s.Assertions.ErrorIs(err, tt.error)
				return
			}

			s.Assertions.Equal(tt.fqmn, fqmn)
		})
	}
}

func (s *FQMNSuite) TestMigrateV1ToV2() {
	for _, tt := range []struct {
		name string
		text string
		hash string
		fqmn FQMN
		error
	}{
		{"migrates v1 to v2", "suborbital.test#default::get-file@v0.0.1", "98qhrfgo3089hafrouhqf48", FQMN{
			Tenant:    "suborbital.test",
			Namespace: "default",
			Name:      "get-file",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
	} {
		s.Run(tt.name, func() {
			fqmn, err := MigrateV1ToV2(tt.text, tt.hash)

			if err != nil {
				s.Assertions.ErrorIs(err, tt.error)
				return
			}

			s.Assertions.Equal(fqmn, tt.fqmn)
		})
	}
}

func (s *FQMNSuite) TestFromParts() {
	for _, tt := range []struct {
		name      string
		namespace string
		ref       string
		tenant    string
		fqmn      string
		error
	}{
		{
			"foobar",
			"default",
			"asdf",
			"com.suborbital.something",
			"fqmn://com.suborbital.something/default/foobar@asdf",
			nil,
		},
		{
			"",
			"default",
			"asdf",
			"com.suborbital.something",
			"",
			ErrFQMNConstructionFailure,
		},
		{
			"foobar",
			"",
			"asdf",
			"com.suborbital.something",
			"",
			ErrFQMNConstructionFailure,
		},
		{
			"foobar",
			"default",
			"",
			"com.suborbital.something",
			"",
			ErrFQMNConstructionFailure,
		},
		{
			"foobar",
			"default",
			"asdf",
			"",
			"",
			ErrFQMNConstructionFailure,
		},
	} {
		s.Run(tt.fqmn, func() {
			fqmn, err := FromParts(tt.tenant, tt.namespace, tt.name, tt.ref)

			if err != nil {
				s.Assertions.ErrorIs(err, tt.error)
				return
			}

			s.Assertions.Equal(fqmn, tt.fqmn)
		})
	}
}
