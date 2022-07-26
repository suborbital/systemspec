package tenant

import (
	"fmt"
	"testing"

	"github.com/suborbital/appspec/tenant/executable"
)

func TestYAMLMarshalUnmarshal(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "user-created",
						},
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
		},
	}

	yamlBytes, err := dir.Marshal()
	if err != nil {
		t.Error(err)
		return
	}

	dir2 := Config{}
	if err := dir2.Unmarshal(yamlBytes); err != nil {
		t.Error(err)
		return
	}

	if err := dir2.Validate(); err != nil {
		t.Error(err)
	}

	if len(dir2.DefaultNamespace.Workflows[0].Steps) != 2 {
		t.Error("wrong number of steps")
		return
	}

	if len(dir2.Modules) != 3 {
		t.Error("wrong number of steps")
		return
	}
}

func TestDirectiveValidatorGroupLast(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}

func TestDirectiveValidatorInvalidOnErr(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}

func TestDirectiveValidatorDuplicateParameterizedResourceMethod(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}

func TestDirectiveValidatorDuplicateResourceMethod(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}

func TestDirectiveValidatorMissingFns(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}

func TestDirectiveFQMNs(t *testing.T) {
	dir := &Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		t.Error("failed to Validate Config")
		return
	}

	run1 := dir.FindModule("getUser")
	if run1 == nil {
		t.Error("failed to find Runnable for getUser")
		return
	}

	FQMN1 := dir.FQMNForFunc(run1.Namespace, run1.Name, "")

	if FQMN1 != "dev.suborbital.appname#default::getUser@v0.1.1" {
		t.Error("FQMN1 should be 'dev.suborbital.appname#default::getUser@v0.1.1', got", FQMN1)
	}

	if FQMN1 != run1.FQMN {
		t.Errorf("FQMN1 %q did not match run1.FQMN %q", FQMN1, run1.FQMN)
	}

	run2 := dir.FindModule("db::getUserDetails")
	if run2 == nil {
		t.Error("failed to find Runnable for db::getUserDetails")
		return
	}

	FQMN2 := dir.FQMNForFunc(run2.Namespace, run2.Name, "")

	if FQMN2 != "dev.suborbital.appname#db::getUserDetails@v0.1.1" {
		t.Error("FQMN2 should be 'dev.suborbital.appname#db::getUserDetails@v0.1.1', got", FQMN2)
	}

	if FQMN2 != run2.FQMN {
		t.Error("FQMN2 did not match run2.FQMN")
	}

	run3 := dir.FindModule("api::returnUser")
	if run3 == nil {
		t.Error("failed to find Runnable for api::returnUser")
		return
	}

	FQMN3 := dir.FQMNForFunc(run3.Namespace, run3.Name, "")

	if FQMN3 != "dev.suborbital.appname#api::returnUser@v0.1.1" {
		t.Error("FQMN3 should be 'dev.suborbital.appname#api::returnUser@v0.1.1', got", FQMN3)
	}

	if FQMN3 != run3.FQMN {
		t.Error("FQMN1 did not match run1.FQMN")
	}

	run4 := dir.FindModule("dev.suborbital.appname#api::returnUser@v0.1.1")
	if run4 == nil {
		t.Error("failed to find Runnable for dev.suborbital.appname#api::returnUser@v0.1.1")
		return
	}

	FQMN4 := dir.FQMNForFunc(run3.Namespace, run3.Name, "")

	if FQMN4 != "dev.suborbital.appname#api::returnUser@v0.1.1" {
		t.Error("FQMN4 should be 'dev.suborbital.appname#api::returnUser@v0.1.1', got", FQMN3)
	}

	if FQMN4 != run4.FQMN {
		t.Error("FQMN1 did not match run1.FQMN")
	}

	run5 := dir.FindModule("foo::bar")
	if run5 != nil {
		t.Error("should not have found a Runnable for foo::bar")
	}
}

func TestDirectiveValidatorWithMissingState(t *testing.T) {
	dir := Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
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
		DefaultNamespace: NamespaceConfig{
			Workflows: []Workflow{
				{
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
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
		},
	}

	if err := dir.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		fmt.Println("Config validation properly failed:", err)
	}
}
