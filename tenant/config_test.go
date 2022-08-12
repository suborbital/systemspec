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
					Name: "Workflow1",
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "user-created",
						},
					},
					Steps: []executable.Executable{
						{
							Group: []executable.ExecutableMod{
								{
									FQMN: "/name/db/getUser",
								},
								{
									FQMN: "/name/db/getUserDetails",
								},
							},
						},
						{
							ExecutableMod: executable.ExecutableMod{
								FQMN: "/name/api/returnUser",
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

func TestConfigValidatorGroupLast(t *testing.T) {
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
					Name: "Workflow1",
					Triggers: []Trigger{
						{
							Source: "nats",
							Topic:  "add-user",
						},
					},
					Steps: []executable.Executable{
						{
							ExecutableMod: executable.ExecutableMod{
								FQMN: "/name/api/returnUser",
							},
						},
						{
							Group: []executable.ExecutableMod{
								{
									FQMN: "/name/db/getUser",
								},
								{
									FQMN: "/name/db/getUserDetails",
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

func TestConfigValidatorMissingFns(t *testing.T) {
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
							Group: []executable.ExecutableMod{
								{
									FQMN: "/name/db/getUser",
								},
								{
									FQMN: "/name/db/getFoobar",
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

func TestConfigFQMNs(t *testing.T) {
	dir := &Config{
		Identifier:    "dev.suborbital.appname",
		TenantVersion: 1,
		Modules: []Module{
			{
				Name:      "getUser",
				Namespace: "default",
				Ref:       "asdf",
			},
			{
				Name:      "getUserDetails",
				Namespace: "db",
				Ref:       "asdf",
			},
			{
				Name:      "returnUser",
				Namespace: "api",
				Ref:       "asdf",
			},
		},
	}

	if err := dir.Validate(); err != nil {
		t.Error("failed to Validate Config")
		return
	}

	mod1, _ := dir.FindModule("/name/default/getUser")
	if mod1 == nil {
		t.Error("failed to FindModule for getUser")
		return
	}

	FQMN1 := dir.FQMNForFunc(mod1.Namespace, mod1.Name, "asdf")

	if FQMN1 != "fqmn://dev.suborbital.appname/asdf/default/getUser" {
		t.Error("FQMN1 should be 'fqmn://dev.suborbital.appname/asdf/default/getUser', got", FQMN1)
	}

	if FQMN1 != mod1.FQMN {
		t.Errorf("FQMN1 %q did not match mod1.FQMN %q", FQMN1, mod1.FQMN)
	}

	mod2, _ := dir.FindModule("/name/db/getUserDetails")
	if mod2 == nil {
		t.Error("failed to FindModule for /name/db/getUserDetails")
		return
	}

	FQMN2 := dir.FQMNForFunc(mod2.Namespace, mod2.Name, "asdf")

	if FQMN2 != "fqmn://dev.suborbital.appname/asdf/db/getUserDetails" {
		t.Error("FQMN2 should be 'fqmn://dev.suborbital.appname/asdf/db/getUserDetails', got", FQMN2)
	}

	if FQMN2 != mod2.FQMN {
		t.Error("FQMN2 did not match mod2.FQMN")
	}

	mod3, _ := dir.FindModule("/name/api/returnUser")
	if mod3 == nil {
		t.Error("failed to FindModule for /name/api/returnUser")
		return
	}

	FQMN3 := dir.FQMNForFunc(mod3.Namespace, mod3.Name, "asdf")

	if FQMN3 != "fqmn://dev.suborbital.appname/asdf/api/returnUser" {
		t.Error("FQMN3 should be 'fqmn://dev.suborbital.appname/asdf/api/returnUser', got", FQMN3)
	}

	if FQMN3 != mod3.FQMN {
		t.Error("FQMN3 did not match mod3.FQMN")
	}

	////////////////////
	// Commented out as the FQMN parser does not currently support /name FQMNs that include the tenant ident
	///////////////////
	// mod4 := dir.FindModule("/name/dev.suborbital.appname/api/returnUser")
	// if mod4 == nil {
	// 	t.Error("failed to FindModule for /name/dev.suborbital.appname/api/returnUser")
	// 	return
	// }

	// FQMN4 := dir.FQMNForFunc(mod3.Namespace, mod3.Name, "asdf")

	// if FQMN4 != "fqmn://dev.suborbital.appname/asdf/api/returnUser" {
	// 	t.Error("FQMN4 should be 'fqmn://dev.suborbital.appname/asdf/api/returnUser', got", FQMN3)
	// }

	// if FQMN4 != mod4.FQMN {
	// 	t.Error("FQMN4 did not match mod4.FQMN")
	// }

	mod5, _ := dir.FindModule("foo::bar")
	if mod5 != nil {
		t.Error("should not have found a Runnable for foo::bar")
	}
}

func TestConfigValidatorWithMissingState(t *testing.T) {
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
							Group: []executable.ExecutableMod{
								{
									FQMN: "/name/db/getUser",
									With: map[string]string{
										"data": "someData",
									},
								},
								{
									FQMN: "/name/db/getFoobar",
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
