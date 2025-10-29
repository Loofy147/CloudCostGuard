package terraform

import (
	"encoding/json"
	"io"
)

// Plan represents the structure of a Terraform plan JSON.
type Plan struct {
	// ResourceChanges is a list of resource changes in the plan.
	ResourceChanges []*ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a change to a single resource in the plan.
type ResourceChange struct {
	// Address is the address of the resource.
	Address      string                 `json:"address"`
	// Type is the type of the resource.
	Type         string                 `json:"type"`
	// Change represents the actions to be taken on a resource.
	Change       Change                 `json:"change"`
	// Before is the state of the resource before the change.
	Before       map[string]interface{} `json:"before"`
	// After is the state of the resource after the change.
	After        map[string]interface{} `json:"after"`
}

// Change represents the actions to be taken on a resource.
type Change struct {
	// Actions is a list of actions to be taken on the resource.
	Actions []string `json:"actions"`
}

// ParsePlan parses a Terraform plan from a JSON reader.
//
// Parameters:
//   r: The io.Reader containing the Terraform plan in JSON format.
//
// Returns:
//   A pointer to the parsed Plan object, or an error if parsing fails.
func ParsePlan(r io.Reader) (*Plan, error) {
	var plan Plan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
