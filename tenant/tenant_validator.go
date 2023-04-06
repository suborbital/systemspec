package tenant

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/suborbital/systemspec/fqmn"
)

// Validate validates a Config.
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
				c.Type != ConnectionTypeKafka) {
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

		c.validateSteps(executableTypeHandler, w.Name, w.Steps, problems)

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

		lastStep := w.Steps[len(w.Steps)-1]
		if w.Response == "" && lastStep.IsGroup() {
			problems.add(fmt.Errorf("workflow for %s has group as last step but does not include 'response' field", w.Name))
		}
	}

	return problems.render()
}

func (c *Config) validateSteps(exType executableType, name string, steps []WorkflowStep, problems *problems) {
	for j, s := range steps {
		if !s.IsSingle() && !s.IsGroup() {
			problems.add(fmt.Errorf("step at position %d for %s %s isn't an Fn or Group", j, exType, name))
		}

		// this function is key as it compartmentalizes 'step validation', and importantly it
		// ensures that a Module is available to handle it and binds it by setting the FQMN field.
		validateFqmn := func(fqmn string) {
			module, err := c.FindModule(fqmn)
			if err != nil {
				problems.add(fmt.Errorf("%s for %s lists mod at step %d that does not have a properly formed FQMN: %s", exType, name, j, fqmn))
			} else if module == nil {
				problems.add(fmt.Errorf("%s for %s lists mod at step %d that does not exist: %s (did you forget a namespace?)", exType, name, j, fqmn))
			}
		}

		// the steps below are referenced by index (j) to ensure the addition of the FQMN in validateFn 'sticks'.
		if s.IsSingle() {
			validateFqmn(steps[j].FQMN)
		} else if s.IsGroup() {
			for p := range s.Group {
				validateFqmn(steps[j].Group[p])
			}
		}
	}
}
