package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRequestIDMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := r.Context().Value(requestIDKey).(string)
		assert.True(t, ok)
		assert.NotEmpty(t, requestID)
	})

	testServer := httptest.NewServer(RequestIDMiddleware()(handler))
	defer testServer.Close()

	http.Get(testServer.URL)
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	testServer := httptest.NewServer(RecoveryMiddleware(zap.NewNop())(handler))
	defer testServer.Close()

	resp, err := http.Get(testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestTimeoutMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(100 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}
	})

	testServer := httptest.NewServer(TimeoutMiddleware(50 * time.Millisecond)(handler))
	defer testServer.Close()

	resp, err := http.Get(testServer.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode)
}
