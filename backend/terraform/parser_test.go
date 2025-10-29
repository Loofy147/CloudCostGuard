package terraform

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestParsePlan(t *testing.T) {
	t.Run("parses a valid plan", func(t *testing.T) {
		planJSON := `
		{
			"resource_changes": [
				{
					"address": "aws_instance.web",
					"type": "aws_instance",
					"change": {
						"actions": ["create"]
					},
					"after": {
						"instance_type": "t2.micro"
					}
				}
			]
		}
		`
		reader := strings.NewReader(planJSON)
		plan, err := ParsePlan(reader)

		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.Len(t, plan.ResourceChanges, 1)

		rc := plan.ResourceChanges[0]
		assert.Equal(t, "aws_instance.web", rc.Address)
		assert.Equal(t, "aws_instance", rc.Type)
		assert.Equal(t, []string{"create"}, rc.Change.Actions)
		assert.Equal(t, "t2.micro", rc.After["instance_type"])
	})

	t.Run("returns error for invalid json", func(t *testing.T) {
		planJSON := `{"invalid_json":}`
		reader := strings.NewReader(planJSON)
		_, err := ParsePlan(reader)
		assert.Error(t, err)
	})
}
