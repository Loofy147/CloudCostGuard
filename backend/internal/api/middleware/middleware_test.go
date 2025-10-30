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

func TestRateLimitMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testServer := httptest.NewServer(RateLimitMiddleware(1, 1)(handler))
	defer testServer.Close()

	resp1, err1 := http.Get(testServer.URL)
	assert.NoError(t, err1)
	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	resp2, err2 := http.Get(testServer.URL)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusTooManyRequests, resp2.StatusCode)
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testServer := httptest.NewServer(APIKeyAuthMiddleware([]string{"test-key"})(handler))
	defer testServer.Close()

	// Test with no key
	req1, _ := http.NewRequest("GET", testServer.URL, nil)
	resp1, err1 := http.DefaultClient.Do(req1)
	assert.NoError(t, err1)
	assert.Equal(t, http.StatusUnauthorized, resp1.StatusCode)

	// Test with invalid key
	req2, _ := http.NewRequest("GET", testServer.URL, nil)
	req2.Header.Set("Authorization", "Bearer invalid-key")
	resp2, err2 := http.DefaultClient.Do(req2)
	assert.NoError(t, err2)
	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode)

	// Test with valid key
	req3, _ := http.NewRequest("GET", testServer.URL, nil)
	req3.Header.Set("Authorization", "Bearer test-key")
	resp3, err3 := http.DefaultClient.Do(req3)
	assert.NoError(t, err3)
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
}
