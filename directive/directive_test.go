package directive

import (
	"fmt"
	"testing"

	"github.com/suborbital/appspec/directive/executable"
)

func TestYAMLMarshalUnmarshal(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []executable.Executable{
					{
						Group: []executable.CallableFn{
							{
								Fn: "db::getUser",
							},
							{
								Fn: "db::getUserDetails",
							},
						},
					},
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
						},
					},
				},
			},
		},
	}

	yamlBytes, err := dir.Marshal()
	if err != nil {
		t.Error(err)
		return
	}

	dir2 := Directive{}
	if err := dir2.Unmarshal(yamlBytes); err != nil {
		t.Error(err)
		return
	}

	if err := dir2.Validate(); err != nil {
		t.Error(err)
	}

	if len(dir2.Handlers[0].Steps) != 2 {
		t.Error("wrong number of steps")
		return
	}

	if len(dir2.Runnables) != 3 {
		t.Error("wrong number of steps")
		return
	}
}

func TestDirectiveValidatorGroupLast(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
						},
					},
					{
						Group: []executable.CallableFn{
							{
								Fn: "db::getUser",
							},
							{
								Fn: "db::getUserDetails",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}

func TestDirectiveValidatorInvalidOnErr(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Code: map[int]string{
									400: "continue",
								},
								Any: "return",
							},
						},
					},
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Other: "continue",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}

func TestDirectiveValidatorDuplicateParameterizedResourceMethod(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/:hello/world",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Any: "continue",
							},
						},
					},
				},
			},
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/:goodbye/moon",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Any: "continue",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}

func TestDirectiveValidatorDuplicateResourceMethod(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/hello",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Any: "continue",
							},
						},
					},
				},
			},
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/hello",
				},
				Steps: []executable.Executable{
					{
						CallableFn: executable.CallableFn{
							Fn: "api::returnUser",
							OnErr: &executable.ErrHandler{
								Any: "continue",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}

func TestDirectiveValidatorMissingFns(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []executable.Executable{
					{
						Group: []executable.CallableFn{
							{
								Fn: "getUser",
							},
							{
								Fn: "getFoobar",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}

func TestDirectiveFQMNs(t *testing.T) {
	dir := &Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "default",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
	}

	if err := dir.Validate(); err != nil {
		t.Error("failed to Validate directive")
		return
	}

	run1 := dir.FindRunnable("getUser")
	if run1 == nil {
		t.Error("failed to find Runnable for getUser")
		return
	}

	FQMN1 := dir.FQMNForFunc(run1.Namespace, run1.Name)

	if FQMN1 != "dev.suborbital.appname#default::getUser@v0.1.1" {
		t.Error("FQMN1 should be 'dev.suborbital.appname#default::getUser@v0.1.1', got", FQMN1)
	}

	if FQMN1 != run1.FQMN {
		t.Errorf("FQMN1 %q did not match run1.FQMN %q", FQMN1, run1.FQMN)
	}

	run2 := dir.FindRunnable("db::getUserDetails")
	if run2 == nil {
		t.Error("failed to find Runnable for db::getUserDetails")
		return
	}

	FQMN2 := dir.FQMNForFunc(run2.Namespace, run2.Name)

	if FQMN2 != "dev.suborbital.appname#db::getUserDetails@v0.1.1" {
		t.Error("FQMN2 should be 'dev.suborbital.appname#db::getUserDetails@v0.1.1', got", FQMN2)
	}

	if FQMN2 != run2.FQMN {
		t.Error("FQMN2 did not match run2.FQMN")
	}

	run3 := dir.FindRunnable("api::returnUser")
	if run3 == nil {
		t.Error("failed to find Runnable for api::returnUser")
		return
	}

	FQMN3 := dir.FQMNForFunc(run3.Namespace, run3.Name)

	if FQMN3 != "dev.suborbital.appname#api::returnUser@v0.1.1" {
		t.Error("FQMN3 should be 'dev.suborbital.appname#api::returnUser@v0.1.1', got", FQMN3)
	}

	if FQMN3 != run3.FQMN {
		t.Error("FQMN1 did not match run1.FQMN")
	}

	run4 := dir.FindRunnable("dev.suborbital.appname#api::returnUser@v0.1.1")
	if run4 == nil {
		t.Error("failed to find Runnable for dev.suborbital.appname#api::returnUser@v0.1.1")
		return
	}

	FQMN4 := dir.FQMNForFunc(run3.Namespace, run3.Name)

	if FQMN4 != "dev.suborbital.appname#api::returnUser@v0.1.1" {
		t.Error("FQMN4 should be 'dev.suborbital.appname#api::returnUser@v0.1.1', got", FQMN3)
	}

	if FQMN4 != run4.FQMN {
		t.Error("FQMN1 did not match run1.FQMN")
	}

	run5 := dir.FindRunnable("foo::bar")
	if run5 != nil {
		t.Error("should not have found a Runnable for foo::bar")
	}
}

func TestDirectiveValidatorWithMissingState(t *testing.T) {
	dir := Directive{
		Identifier:  "dev.suborbital.appname",
		AppVersion:  "v0.1.1",
		AtmoVersion: "v0.0.6",
		Runnables: []Runnable{
			{
				Name:      "getUser",
				Namespace: "db",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
			},
		},
		Handlers: []Handler{
			{
				Input: Input{
					Type:     "request",
					Method:   "GET",
					Resource: "/api/v1/user",
				},
				Steps: []executable.Executable{
					{
						Group: []executable.CallableFn{
							{
								Fn: "getUser",
								With: map[string]string{
									"data": "someData",
								},
							},
							{
								Fn: "getFoobar",
							},
						},
					},
				},
			},
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("directive validation should have failed")
	} else {
		fmt.Println("directive validation properly failed:", err)
	}
}
