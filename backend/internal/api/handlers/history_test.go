package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestHistoryHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"pr_number", "total_monthly_cost", "created_at"}).
		AddRow(123, 123.45, time.Now())
	mock.ExpectQuery("SELECT pr_number, total_monthly_cost, created_at FROM estimations WHERE repository = \\$1 ORDER BY created_at DESC").
		WithArgs("test-owner/test-repo").
		WillReturnRows(rows)

	handler := NewHistoryHandler(db, zap.NewNop())

	req, _ := http.NewRequest("GET", "/history/test-owner/test-repo", nil)
	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Handle("/history/{owner}/{repo}", handler)
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var estimations []struct {
		PRNumber        int       `json:"pr_number"`
		TotalMonthlyCost float64   `json:"total_monthly_cost"`
		CreatedAt       time.Time `json:"created_at"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &estimations)
	assert.NoError(t, err)
	assert.Len(t, estimations, 1)
	assert.Equal(t, 123, estimations[0].PRNumber)
}
