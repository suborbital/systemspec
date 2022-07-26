package tenant

import (
	"fmt"

	"github.com/pkg/errors"

	fqmn "github.com/suborbital/appspec/fqfn"
	"github.com/suborbital/appspec/tenant/executable"
)

// Validate validates a directive
func (c *Config) Validate() (err error) {
	problems := &problems{}

	c.calculateFQMNs()

	if c.Identifier == "" {
		problems.add(errors.New("identifier is missing"))
	}

	if len(c.Modules) < 1 {
		problems.add(errors.New("no modules listed"))
	}

	fns := map[string]bool{}

	for i, f := range c.Modules {
		namespaced := fmt.Sprintf("%s::%s", f.Namespace, f.Name)

		if _, exists := fns[namespaced]; exists {
			problems.add(fmt.Errorf("duplicate fn %s found", namespaced))
			continue
		}

		if _, exists := fns[f.Name]; exists {
			problems.add(fmt.Errorf("duplicate fn %s found", namespaced))
			continue
		}

		if f.Name == "" {
			problems.add(fmt.Errorf("function at position %d missing name", i))
			continue
		}

		if f.Namespace == "" {
			problems.add(fmt.Errorf("function at position %d missing namespace", i))
		}

		// if the fn is in the default namespace, let it exist "naked" and namespaced.
		if f.Namespace == fqmn.NamespaceDefault {
			fns[f.Name] = true
			fns[namespaced] = true
		} else {
			fns[namespaced] = true
		}
	}

	if err := c.validateNamespaceConfig(c.DefaultNamespace); err != nil {
		problems.add(err)
	}

	return problems.render()
}

type executableType string

const (
	executableTypeHandler  = executableType("handler")
	executableTypeSchedule = executableType("schedule")
)

func (c *Config) validateNamespaceConfig(nc NamespaceConfig) (err error) {
	problems := &problems{}

	// validate connections before handlers because we want to make sure they're all correct first.
	if nc.Connections != nil && len(nc.Connections) > 0 {
		for _, c := range nc.Connections {
			if c.Type == "" || (c.Type != ConnectionTypeNATS &&
				c.Type != ConnectionTypeKafka &&
				c.Type != ConnectionTypeRedis &&
				c.Type != ConnectionTypeMySQL &&
				c.Type != ConnectionTypePostgres) {
				problems.add(fmt.Errorf("unknown connection type %s", c.Type))
			}
		}

	}

	if nc.Authentication != nil {
		if nc.Authentication.Domains != nil {
			for d, h := range nc.Authentication.Domains {
				if h.HeaderType == "" {
					h.HeaderType = "bearer"
				}

				if h.Value == "" {
					problems.add(fmt.Errorf("authentication for domain %s has an empty value", d))
				}
			}
		}
	}

	// Conflicting routes will result in a panic which we catch here
	defer func() {
		if r := recover(); r != nil {
			problems.add(fmt.Errorf("%s", r))
		}
		err = problems.render()
	}()

	uniqueWorkflowNames := map[string]struct{}{}

	for i, w := range nc.Workflows {
		if w.Name == "" {
			problems.add(fmt.Errorf("workflow at position %d has no name", i))
			continue
		} else {
			if _, exists := uniqueWorkflowNames[w.Name]; exists {
				problems.add(fmt.Errorf("workflow at position %d has a non-unique name %s", i, w.Name))
			}

			uniqueWorkflowNames[w.Name] = struct{}{}
		}

		if len(w.Steps) == 0 {
			problems.add(fmt.Errorf("workflow %s missing steps", w.Name))
			continue
		}

		c.validateSteps(executableTypeHandler, w.Name, w.Steps, map[string]bool{}, problems)

		if w.Schedule != nil {
			if w.Schedule.Every.Seconds == 0 && w.Schedule.Every.Minutes == 0 && w.Schedule.Every.Hours == 0 && w.Schedule.Every.Days == 0 {
				problems.add(fmt.Errorf("workflow %s's schedule has no 'every' values", w.Name))
			}

			// user can provide an 'initial state' via the schedule.State field, so let's prime the state with it.
			initialState := map[string]bool{}
			for k := range w.Schedule.State {
				initialState[k] = true
			}
		}
	}

	for i, q := range nc.Queries {
		if q.Name == "" {
			problems.add(fmt.Errorf("query at position %d has no name", i))
		}

		if q.Query == "" {
			problems.add(fmt.Errorf("query at position %d has no query value", i))
		}

		if q.Type != "" {
			if q.Type != queryTypeInsert && q.Type != queryTypeSelect && q.Type != queryTypeUpdate && q.Type != queryTypeDelete {
				problems.add(fmt.Errorf("query at position %d has invalid type %s", i, q.Type))
			}
		}

		if q.VarCount < 0 {
			problems.add(fmt.Errorf("query at position %d cannot have negative var count", i))
		}
	}

	return problems.render()
}

func (c *Config) validateSteps(exType executableType, name string, steps []executable.Executable, initialState map[string]bool, problems *problems) map[string]bool {
	// keep track of the functions that have run so far at each step.
	fullState := initialState

	for j, s := range steps {
		fnsToAdd := []string{}

		if !s.IsFn() && !s.IsGroup() {
			if s.ForEach != nil {
				problems.add(fmt.Errorf("step at position %d for %s %s is a 'forEach', which was removed in v0.4.0", j, exType, name))
			} else {
				problems.add(fmt.Errorf("step at position %d for %s %s isn't an Fn or Group", j, exType, name))
			}
		}

		// this function is key as it compartmentalizes 'step validation', and importantly it
		// ensures that a Runnable is available to handle it and binds it by setting the FQFN field.
		validateFn := func(fn *executable.CallableFn) {
			runnable := c.FindModule(fn.Fn)
			if runnable == nil {
				problems.add(fmt.Errorf("%s for %s lists fn at step %d that does not exist: %s (did you forget a namespace?)", exType, name, j, fn.Fn))
			} else {
				fn.FQFN = runnable.FQMN
			}

			for _, key := range fn.With {
				if _, exists := fullState[key]; !exists {
					problems.add(fmt.Errorf("%s for %s has 'with' value at step %d referencing a key that is not yet available in the handler's state: %s", exType, name, j, key))
				}
			}

			if fn.OnErr != nil {
				// if codes are specificed, 'other' should be used, not 'any'.
				if len(fn.OnErr.Code) > 0 && fn.OnErr.Any != "" {
					problems.add(fmt.Errorf("%s for %s has 'onErr.any' value at step %d while specific codes are specified, use 'other' instead", exType, name, j))
				} else if fn.OnErr.Any != "" {
					if fn.OnErr.Any != "continue" && fn.OnErr.Any != "return" {
						problems.add(fmt.Errorf("%s for %s has 'onErr.any' value at step %d with an invalid error directive: %s", exType, name, j, fn.OnErr.Any))
					}
				}

				// if codes are NOT specificed, 'any' should be used, not 'other'.
				if len(fn.OnErr.Code) == 0 && fn.OnErr.Other != "" {
					problems.add(fmt.Errorf("%s for %s has 'onErr.other' value at step %d while specific codes are not specified, use 'any' instead", exType, name, j))
				} else if fn.OnErr.Other != "" {
					if fn.OnErr.Other != "continue" && fn.OnErr.Other != "return" {
						problems.add(fmt.Errorf("%s for %s has 'onErr.any' value at step %d with an invalid error directive: %s", exType, name, j, fn.OnErr.Other))
					}
				}

				for code, val := range fn.OnErr.Code {
					if val != "return" && val != "continue" {
						problems.add(fmt.Errorf("%s for %s has 'onErr.code' value at step %d with an invalid error directive for code %d: %s", exType, name, j, code, val))
					}
				}
			}

			key := fn.Fn
			if fn.As != "" {
				key = fn.As
			}

			fnsToAdd = append(fnsToAdd, key)
		}

		// the steps below are referenced by index (j) to ensure the addition of the FQFN in validateFn 'sticks'.
		if s.IsFn() {
			validateFn(&steps[j].CallableFn)
		} else if s.IsGroup() {
			for p := range s.Group {
				validateFn(&steps[j].Group[p])
			}
		}

		for _, newFn := range fnsToAdd {
			fullState[newFn] = true
		}
	}

	return fullState
}
