package cloudconnexa

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// fastRetries shrinks the backoff bounds for the duration of a test.
func fastRetries(t *testing.T, base, maxWait time.Duration) {
	t.Helper()
	oldBase, oldMax := retryBaseWait, retryMaxWait
	retryBaseWait, retryMaxWait = base, maxWait
	t.Cleanup(func() { retryBaseWait, retryMaxWait = oldBase, oldMax })
}

// newRetryTestClient builds a client without rate limiters, as a caller constructing
// the struct directly might, so the nil-limiter path is exercised too.
func newRetryTestClient(server *httptest.Server, maxRetries int) *Client {
	return &Client{
		client:     server.Client(),
		BaseURL:    server.URL,
		Token:      "mock-access-token",
		maxRetries: maxRetries,
	}
}

// TestDoRequest_RetriesOn429ThenSucceeds verifies that a request rejected with 429 is retried
// with the same body and headers until it succeeds.
func TestDoRequest_RetriesOn429ThenSucceeds(t *testing.T) {
	fastRetries(t, time.Millisecond, 10*time.Millisecond)

	var mu sync.Mutex
	var bodies, auths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		mu.Lock()
		bodies = append(bodies, string(body))
		auths = append(auths, r.Header.Get("Authorization"))
		n := len(bodies)
		mu.Unlock()
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/things", bytes.NewBuffer([]byte(`{"name":"x"}`)))
	require.NoError(t, err)

	body, err := client.DoRequest(req)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(body))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, []string{`{"name":"x"}`, `{"name":"x"}`, `{"name":"x"}`}, bodies, "body must be replayed on every attempt")
	for _, auth := range auths {
		assert.Equal(t, "Bearer mock-access-token", auth)
	}
}

// TestDoRequest_GivesUpAfterMaxRetries verifies that the last 429 is returned as an
// *ErrClientResponse once the retry budget is spent, for both explicit and default budgets.
func TestDoRequest_GivesUpAfterMaxRetries(t *testing.T) {
	fastRetries(t, time.Millisecond, 10*time.Millisecond)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	tests := []struct {
		name       string
		maxRetries int
		wantCalls  int32
	}{
		{"explicit budget", 2, 3},
		{"zero means default", 0, DefaultMaxRetries + 1},
		{"negative disables retry", -1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls.Store(0)
			client := newRetryTestClient(server, tt.maxRetries)
			req, _ := http.NewRequest(http.MethodGet, server.URL+"/things", nil)

			_, err := client.DoRequest(req)

			var apiErr *ErrClientResponse
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode())
			assert.Equal(t, `{"error":"rate limited"}`, apiErr.Body())
			assert.Equal(t, tt.wantCalls, calls.Load())
		})
	}
}

// TestDoRequest_DoesNotRetryOtherStatuses verifies that only 429 is retried, even when the
// response carries a Retry-After header.
func TestDoRequest_DoesNotRetryOtherStatuses(t *testing.T) {
	fastRetries(t, time.Millisecond, 10*time.Millisecond)

	for _, status := range []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError, http.StatusServiceUnavailable} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(status)
			}))
			defer server.Close()

			client := newRetryTestClient(server, 0)
			req, _ := http.NewRequest(http.MethodPost, server.URL+"/things", bytes.NewBufferString(`{}`))

			_, err := client.DoRequest(req)

			var apiErr *ErrClientResponse
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, status, apiErr.StatusCode())
			assert.Equal(t, int32(1), calls.Load())
		})
	}
}

// TestDoRequest_HonorsRetryAfter verifies that the retry waits as long as the server asks.
func TestDoRequest_HonorsRetryAfter(t *testing.T) {
	fastRetries(t, time.Millisecond, 5*time.Second)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/things", nil)

	start := time.Now()
	_, err := client.DoRequest(req)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, elapsed, time.Second)
	assert.Less(t, elapsed, 3*time.Second)
	assert.Equal(t, int32(2), calls.Load())
}

// TestDoRequest_ContextCancelledDuringBackoff verifies that a wait between retries ends as
// soon as the request context is done.
func TestDoRequest_ContextCancelledDuringBackoff(t *testing.T) {
	fastRetries(t, time.Millisecond, 30*time.Second)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	client := newRetryTestClient(server, 0)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/things", nil)

	start := time.Now()
	_, err := client.DoRequest(req)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 2*time.Second)
	assert.Equal(t, int32(1), calls.Load())
}

// TestDoRequest_SharedCooldownAcrossRequests verifies that a request issued while another
// request on the same client is backing off waits for the same deadline instead of hitting
// the server straight away.
func TestDoRequest_SharedCooldownAcrossRequests(t *testing.T) {
	fastRetries(t, time.Millisecond, 5*time.Second)

	var mu sync.Mutex
	var times []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		times = append(times, time.Now())
		n := len(times)
		mu.Unlock()
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/a", nil)
		_, err := client.DoRequest(req)
		assert.NoError(t, err)
	}()

	// Wait until the first request has been rejected and the pause is in place.
	require.Eventually(t, func() bool { return !client.pauseRead.deadline().IsZero() }, 5*time.Second, time.Millisecond)

	req, _ := http.NewRequest(http.MethodGet, server.URL+"/b", nil)
	_, err := client.DoRequest(req)
	require.NoError(t, err)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, times, 3, "one rejected request, one retry, one second request")
	for _, later := range times[1:] {
		assert.GreaterOrEqual(t, later.Sub(times[0]), time.Second)
	}
}

// opaqueReader hides the concrete reader type so http.NewRequest cannot populate GetBody.
type opaqueReader struct{ io.Reader }

// TestDoRequest_NoRetryWhenBodyCannotBeReplayed verifies that a request whose body cannot
// be re-read is never replayed, since a retry would send an empty body.
func TestDoRequest_NoRetryWhenBodyCannotBeReplayed(t *testing.T) {
	fastRetries(t, time.Millisecond, 10*time.Millisecond)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/things", opaqueReader{strings.NewReader(`{}`)})
	require.NoError(t, err)
	require.Nil(t, req.GetBody)

	_, err = client.DoRequest(req)

	var apiErr *ErrClientResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode())
	assert.Equal(t, int32(1), calls.Load())
}

// TestDoRequest_OptionalLimiterStillThrottles verifies that a caller who sets a rate limiter
// still gets proactive pacing on top of the retry.
func TestDoRequest_OptionalLimiterStillThrottles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	client.UpdateRateLimiter = rate.NewLimiter(rate.Every(200*time.Millisecond), 1)

	start := time.Now()
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/things", bytes.NewBufferString(`{}`))
		_, err := client.DoRequest(req)
		require.NoError(t, err)
	}

	assert.GreaterOrEqual(t, time.Since(start), 400*time.Millisecond)
}

// TestRetryWait checks how server hints and the jittered backoff combine into a wait.
func TestRetryWait(t *testing.T) {
	fastRetries(t, 100*time.Millisecond, 2*time.Second)

	header := func(kv ...string) http.Header {
		h := http.Header{}
		for i := 0; i < len(kv); i += 2 {
			h.Set(kv[i], kv[i+1])
		}
		return h
	}
	httpDate := func(d time.Duration) string { return time.Now().Add(d).UTC().Format(http.TimeFormat) }

	tests := []struct {
		name     string
		header   http.Header
		attempt  int
		min, max time.Duration
	}{
		{"no hints, first retry is jittered base", header(), 0, 50 * time.Millisecond, 100 * time.Millisecond},
		{"no hints, backoff doubles per attempt", header(), 3, 400 * time.Millisecond, 800 * time.Millisecond},
		{"no hints, backoff is capped", header(), 20, time.Second, 2 * time.Second},
		{"Retry-After seconds wins over a smaller backoff", header("Retry-After", "1"), 0, time.Second, time.Second},
		{"Retry-After is capped", header("Retry-After", "600"), 0, 2 * time.Second, 2 * time.Second},
		{"Retry-After HTTP-date", header("Retry-After", httpDate(2*time.Second)), 0, 900 * time.Millisecond, 2 * time.Second},
		{"Retry-After in the past falls back to backoff", header("Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT"), 0, 50 * time.Millisecond, 100 * time.Millisecond},
		{"replenish headers give the time per token", header("X-RateLimit-Replenish-Rate", "10", "X-RateLimit-Replenish-Time", "15"), 0, 1500 * time.Millisecond, 1500 * time.Millisecond},
		{"backoff wins over a smaller hint", header("Retry-After", "0"), 3, 400 * time.Millisecond, 800 * time.Millisecond},
		{"malformed hints fall back to backoff", header("Retry-After", "soon", "X-RateLimit-Replenish-Rate", "0", "X-RateLimit-Replenish-Time", "x"), 0, 50 * time.Millisecond, 100 * time.Millisecond},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryWait(tt.header, tt.attempt)
			assert.GreaterOrEqual(t, got, tt.min)
			assert.LessOrEqual(t, got, tt.max)
		})
	}
}

// TestNewClientWithOptions_RetryDefaults verifies the defaults a constructed client starts with:
// no proactive throttling and the retry budget taken from the options.
func TestNewClientWithOptions_RetryDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok"}`))
	}))
	defer server.Close()

	c, err := NewClientWithOptions(server.URL, "id", "secret", &ClientOptions{AllowInsecureHTTP: true})
	require.NoError(t, err)
	assert.Equal(t, 0, c.maxRetries, "zero is resolved to DefaultMaxRetries per request")
	assert.Equal(t, rate.Inf, c.ReadRateLimiter.Limit(), "no proactive throttling by default")
	assert.Equal(t, rate.Inf, c.UpdateRateLimiter.Limit(), "no proactive throttling by default")

	c, err = NewClientWithOptions(server.URL, "id", "secret", &ClientOptions{AllowInsecureHTTP: true, MaxRetries: 2})
	require.NoError(t, err)
	assert.Equal(t, 2, c.maxRetries)

	c, err = NewClientWithOptions(server.URL, "id", "secret", &ClientOptions{AllowInsecureHTTP: true, MaxRetries: -1})
	require.NoError(t, err)
	assert.Equal(t, -1, c.maxRetries, "negative is kept so DoRequest disables retry")
	assert.Nil(t, c.onRetry)

	hook := func(*http.Request, int, time.Duration) {}
	c, err = NewClientWithOptions(server.URL, "id", "secret", &ClientOptions{AllowInsecureHTTP: true, OnRetry: hook})
	require.NoError(t, err)
	assert.NotNil(t, c.onRetry)
}

// TestDoRequest_OnRetryHook verifies that the hook is called once per retry with the retry
// number and the wait, and not at all for the final failure or for other statuses.
func TestDoRequest_OnRetryHook(t *testing.T) {
	fastRetries(t, time.Millisecond, 5*time.Second)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) <= 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	type call struct {
		method, path string
		attempt      int
		wait         time.Duration
	}
	var mu sync.Mutex
	var got []call
	client := newRetryTestClient(server, 0)
	client.onRetry = func(req *http.Request, attempt int, wait time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, call{req.Method, req.URL.Path, attempt, wait})
	}

	req, _ := http.NewRequest(http.MethodPost, server.URL+"/things", bytes.NewBufferString(`{}`))
	_, err := client.DoRequest(req)

	var apiErr *ErrClientResponse
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode())

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, got, 2, "one call per retry, none for the 400")
	for i, c := range got {
		assert.Equal(t, http.MethodPost, c.method)
		assert.Equal(t, "/things", c.path)
		assert.Equal(t, i+1, c.attempt)
		assert.InDelta(t, time.Second, c.wait, float64(100*time.Millisecond))
	}
}

// TestDoRequest_RetriesAreSpacedApart verifies that concurrent requests rejected together are
// retried one server hint apart rather than all at the same instant, so each retry finds a
// replenished token instead of racing the others for it.
func TestDoRequest_RetriesAreSpacedApart(t *testing.T) {
	fastRetries(t, time.Millisecond, 5*time.Second)
	const workers = 4
	const hint = 200 * time.Millisecond // X-RateLimit-Replenish-Rate 5 per 1 s

	var mu sync.Mutex
	var rejected int
	var retries []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if rejected < workers {
			rejected++
			w.Header().Set("X-RateLimit-Replenish-Rate", "5")
			w.Header().Set("X-RateLimit-Replenish-Time", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		retries = append(retries, time.Now())
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, server.URL+"/things", nil)
			_, err := client.DoRequest(req)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, retries, workers, "every request succeeds on its first retry")
	for i := 1; i < len(retries); i++ {
		gap := retries[i].Sub(retries[i-1])
		assert.GreaterOrEqual(t, gap, hint-20*time.Millisecond, "retry %d fired %s after retry %d", i, gap, i-1)
	}
}

// TestDoRequest_ReadAndWritePausesAreIndependent verifies that a 429 on a write does not hold
// up reads, since the API limits the two in separate buckets.
func TestDoRequest_ReadAndWritePausesAreIndependent(t *testing.T) {
	fastRetries(t, time.Millisecond, 5*time.Second)

	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && posts.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := newRetryTestClient(server, 0)
	go func() {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/things", bytes.NewBufferString(`{}`))
		_, _ = client.DoRequest(req)
	}()
	require.Eventually(t, func() bool { return !client.pauseWrite.deadline().IsZero() }, 5*time.Second, time.Millisecond)

	start := time.Now()
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/things", nil)
	_, err := client.DoRequest(req)
	require.NoError(t, err)
	assert.Less(t, time.Since(start), 500*time.Millisecond, "GET must not wait for the write pause")
}
