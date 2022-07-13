package fqmn

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type FQFNSuite struct {
	suite.Suite
}

func TestFQFNSuite(t *testing.T) {
	suite.Run(t, &FQFNSuite{})
}

func (s *FQFNSuite) TestParse() {
	for _, tt := range []struct {
		name string
		text string
		fqmn FQMN
		error
	}{
		{"fully-qualified single-level namespace", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api-users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api-users",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"fully-qualified two-level namespace", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"fully-qualified multi-level namespace", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/auora/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users/auora",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for single-application func with single-level namespace", "/api-users/add-user", FQMN{
			Namespace: "api-users",
			Fn:        "add-user",
		}, nil},
		{"uri for single-application func with two-level namespace", "/api/users/add-user", FQMN{
			Namespace: "api/users",
			Fn:        "add-user",
		}, nil},
		{"uri for single-application func with multi-level namespace", "/api/users/auora/add-user", FQMN{
			Namespace: "api/users/auora",
			Fn:        "add-user",
		}, nil},
		{"uri for versioned single-application func with single-level namespace", "/ref/98qhrfgo3089hafrouhqf48/api-users/add-user", FQMN{
			Namespace: "api-users",
			Fn:        "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for versioned single-application func with two-level namespace", "/ref/98qhrfgo3089hafrouhqf48/api/users/add-user", FQMN{
			Namespace: "api/users",
			Fn:        "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for versioned single-application func with multi-level namespace", "/ref/98qhrfgo3089hafrouhqf48/api/users/auora/add-user", FQMN{
			Namespace: "api/users/auora",
			Fn:        "add-user",
			Ref:       "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for multi-application func with single-level namespace", "/com.suborbital.acmeco/api-users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api-users",
			Fn:         "add-user",
		}, nil},
		{"uri for multi-application func with two-level namespace", "/com.suborbital.acmeco/api/users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users",
			Fn:         "add-user",
		}, nil},
		{"uri for multi-application func with multi-level namespace", "/com.suborbital.acmeco/api/users/auora/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users/auora",
			Fn:         "add-user",
		}, nil},
		{"uri for versioned multi-application func with single-level namespace", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco/api-users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api-users",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for versioned multi-application func with two-level namespace", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco/api/users/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"uri for versioned multi-application func with multi-level namespace", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco/api/users/auora/add-user", FQMN{
			Identifier: "com.suborbital.acmeco",
			Namespace:  "api/users/auora",
			Fn:         "add-user",
			Ref:        "98qhrfgo3089hafrouhqf48",
		}, nil},
		{"malformed—doesn't start with right prefix 1", "fqmn:com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/auora/add-user", FQMN{}, errWrongPrefix},
		{"malformed—doesn't start with right prefix 2", "https://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/auora/add-user", FQMN{}, errWrongPrefix},
		{"malformed—not fully-qualified 1", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48", FQMN{}, errMustBeFullyQualified},
		{"malformed—not fully-qualified 2", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/add-user", FQMN{}, errMustBeFullyQualified},
		{"malformed—not enough parts 1", "/add-user", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 2", "/com.suborbital.acmeco", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 3", "/com.suborbital.acmeco/add-user", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 4", "/ref/98qhrfgo3089hafrouhqf48", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 5", "/ref/98qhrfgo3089hafrouhqf48/add-user", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 6", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco", FQMN{}, errTooFewParts},
		{"malformed—not enough parts 7", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco/add-user", FQMN{}, errTooFewParts},
		{"malformed—trailing slash 1", "fqmn://com.suborbital.acmeco/98qhrfgo3089hafrouhqf48/api/users/add-user/", FQMN{}, errTrailingSlash},
		{"malformed—trailing slash 2", "/ref/98qhrfgo3089hafrouhqf48/com.suborbital.acmeco/api/users/auora/add-user/", FQMN{}, errTrailingSlash},
	} {
		s.Run(tt.name, func() {
			fqmn, err := Parse(tt.text)

			if err != nil {
				s.Assertions.ErrorIs(err, tt.error)
				return
			}

			s.Assertions.Equal(fqmn, tt.fqmn)
		})
	}
}

func (s *FQFNSuite) TestMigrateV1ToV2() {
	for _, tt := range []struct {
		name string
		text string
		hash string
		fqmn FQMN
		error
	}{
		{"migrates v1 to v2", "com.suborbital.test#default::get-file@v0.0.1", "98qhrfgo3089hafrouhqf48", FQMN{
			Identifier: "com.suborbital.test",
			Namespace:  "default",
			Fn:         "get-file",
			Ref:        "98qhrfgo3089hafrouhqf48",
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
