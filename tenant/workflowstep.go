package tenant

import (
	"errors"
)

var (
	// ErrSequenceShouldReturn is represents a failed function call that should result in a return.
	ErrSequenceShouldReturn = errors.New("function resulted in a Run Error and sequence should return")
	ErrSequenceCompleted    = errors.New("sequence is complete, no steps to run")
)

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	FQMN  string   `yaml:"fqmn" json:"fqmn"`
	Group []string `yaml:"group,omitempty" json:"group,omitempty"`
}

// IsGroup returns true if the WorkflowStep is a group.
func (e WorkflowStep) IsGroup() bool {
	return e.FQMN == "" && e.Group != nil && len(e.Group) > 0
}

// IsSingle returns true if the WorkflowStep is a group.
func (e WorkflowStep) IsSingle() bool {
	return e.FQMN != "" && e.Group == nil
}
