package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudcostguard/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEstimateHandler_InputValidation(t *testing.T) {
	estimatorSvc := service.NewEstimator(nil, zap.NewNop())
	handler := NewEstimateHandler(estimatorSvc, zap.NewNop())

	// Test with nil plan
	req1, _ := http.NewRequest("POST", "/estimate", bytes.NewBufferString(`{"plan": null}`))
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusBadRequest, rr1.Code)
	assert.Contains(t, rr1.Body.String(), "plan cannot be nil")

	// Test with too many resources
	resources := strings.Repeat(`{"address": "a"},`, 1001)
	resources = strings.TrimRight(resources, ",")
	req2, _ := http.NewRequest("POST", "/estimate", bytes.NewBufferString(`{"plan": {"resource_changes": [`+resources+`]}}`))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
	assert.Contains(t, rr2.Body.String(), "too many resources in plan")

	// Test with empty resource address
	req3, _ := http.NewRequest("POST", "/estimate", bytes.NewBufferString(`{"plan": {"resource_changes": [{"address": ""}]}}`))
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusBadRequest, rr3.Code)
	assert.Contains(t, rr3.Body.String(), "resource address cannot be empty")
}
