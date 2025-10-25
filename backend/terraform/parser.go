package terraform

import (
	"encoding/json"
	"io"
)

// Plan represents the structure of a Terraform plan JSON.
type Plan struct {
	ResourceChanges []*ResourceChange `json:"resource_changes"`
}

// ResourceChange represents a change to a single resource in the plan.
type ResourceChange struct {
	Address      string                 `json:"address"`
	Type         string                 `json:"type"`
	Change       Change                 `json:"change"`
	Before       map[string]interface{} `json:"before"`
	After        map[string]interface{} `json:"after"`
}

// Change represents the actions to be taken on a resource.
type Change struct {
	Actions []string `json:"actions"`
}

// ParsePlan parses a Terraform plan from a JSON reader.
func ParsePlan(r io.Reader) (*Plan, error) {
	var plan Plan
	if err := json.NewDecoder(r).Decode(&plan); err != nil {
		return nil, err
	}
	return &plan, nil
}
