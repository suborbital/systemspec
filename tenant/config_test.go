package tenant

import (
	"fmt"
	"testing"
)

func TestYAMLMarshalUnmarshal(t *testing.T) {
	conf := Config{
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
					Steps: []WorkflowStep{
						{
							Group: []string{"/name/db/getUser", "/name/db/getUserDetails"},
						},
						{
							FQMN: "/name/api/returnUser",
						},
					},
				},
			},
		},
	}

	yamlBytes, err := conf.Marshal()
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
	conf := Config{
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
					Steps: []WorkflowStep{
						{
							FQMN: "/name/api/returnUser",
						},
						{
							Group: []string{"/name/db/getUser", "/name/db/getUserDetails"},
						},
					},
				},
			},
		},
	}

	if err := conf.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		_, _ = fmt.Println("Config validation properly failed:", err)
	}
}

func TestConfigValidatorMissingFns(t *testing.T) {
	conf := Config{
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
					Steps: []WorkflowStep{
						{
							Group: []string{"/name/db/getUser", "/name/db/getFoobar"},
						},
					},
				},
			},
		},
	}

	if err := conf.Validate(); err == nil {
		t.Error("Config validation should have failed")
	} else {
		_, _ = fmt.Println("Config validation properly failed:", err)
	}
}

func TestConfigFQMNs(t *testing.T) {
	conf := &Config{
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

	if err := conf.Validate(); err != nil {
		t.Error("failed to Validate Config")
		return
	}

	mod1, _ := conf.FindModule("/name/default/getUser")
	if mod1 == nil {
		t.Error("failed to FindModule for getUser")
		return
	}

	FQMN1, _ := conf.FQMNForFunc(mod1.Namespace, mod1.Name, "asdf")

	if FQMN1 != "fqmn://dev.suborbital.appname/default/getUser@asdf" {
		t.Error("FQMN1 should be 'fqmn://dev.suborbital.appname/default/getUser@asdf', got", FQMN1)
	}

	if FQMN1 != mod1.FQMN {
		t.Errorf("FQMN1 %q did not match mod1.FQMN %q", FQMN1, mod1.FQMN)
	}

	mod2, _ := conf.FindModule("/name/db/getUserDetails")
	if mod2 == nil {
		t.Error("failed to FindModule for /name/db/getUserDetails")
		return
	}

	FQMN2, _ := conf.FQMNForFunc(mod2.Namespace, mod2.Name, "asdf")

	if FQMN2 != "fqmn://dev.suborbital.appname/db/getUserDetails@asdf" {
		t.Error("FQMN2 should be 'fqmn://dev.suborbital.appname/db/getUserDetails@asdf', got", FQMN2)
	}

	if FQMN2 != mod2.FQMN {
		t.Error("FQMN2 did not match mod2.FQMN")
	}

	mod3, _ := conf.FindModule("/name/api/returnUser")
	if mod3 == nil {
		t.Error("failed to FindModule for /name/api/returnUser")
		return
	}

	FQMN3, _ := conf.FQMNForFunc(mod3.Namespace, mod3.Name, "asdf")

	if FQMN3 != "fqmn://dev.suborbital.appname/api/returnUser@asdf" {
		t.Error("FQMN3 should be 'fqmn://dev.suborbital.appname/api/returnUser@asdf', got", FQMN3)
	}

	if FQMN3 != mod3.FQMN {
		t.Error("FQMN3 did not match mod3.FQMN")
	}

	mod5, _ := conf.FindModule("foo::bar")
	if mod5 != nil {
		t.Error("should not have found a Module for foo::bar")
	}
}
