// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package plextrac_test

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/brimstone/plextraccli/plextrac"
)

func genJWT(exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload := map[string]any{"exp": exp.Unix()}

	pBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	encoded := strings.TrimRight(base64.RawURLEncoding.EncodeToString(pBytes), "=")

	return header + "." + encoded + ".fakesig"
}

func testServerWithHandler(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	t.Helper()

	server := httptest.NewTLSServer(handler)

	return server, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test server only
		},
	}
}

func TestNew_accepts_custom_httpclient(t *testing.T) {
	t.Parallel()

	custom := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test server only
		},
	}

	refreshCount := 0

	var mu sync.Mutex

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "token/refresh") {
			refreshCount++

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":    "success",
				"tenant_id": 42,
				"token":     genJWT(time.Now().Add(24 * time.Hour)),
			})
			if err != nil {
				t.Fail()
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{{"id": "abc123", "name": "demo"}},
		})
		if err != nil {
			t.Fail()
		}
	}

	server, _ := testServerWithHandler(t, handler)

	defer server.Close()

	expiredSoon := time.Now().Add(60 * time.Second)

	o := plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWT(expiredSoon),

		HTTPClient: custom,
	}

	ua, _, err := plextrac.New(o)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	ua.GetTenantID()

	_, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if custom.Timeout != 5*time.Second {
		t.Fatal("expected custom httpClient to have 5s timeout")
	}
}

func TestNew_defaults_to_http_defaultclient(t *testing.T) {
	t.Parallel()

	o := plextrac.NewOptions{
		InstanceURL: "example.com",
		AuthToken:   genJWT(time.Now().Add(24 * time.Hour)),
	}

	_, _, err := plextrac.New(o)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
}

func TestLogin_uses_httpclient(t *testing.T) {
	t.Parallel()

	var (
		loginCallCount int
		mu             sync.Mutex
		newToken       string
		savedToken     string
		savedExpires   time.Time
	)

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		loginCallCount++

		w.Header().Set("Content-Type", "application/json")

		newToken = genJWT(time.Now().Add(24 * time.Hour))

		err := json.NewEncoder(w).Encode(map[string]any{
			"status":    "success",
			"tenant_id": 42,
			"token":     newToken,
		})
		if err != nil {
			t.Fail()
		}
	}

	server, client := testServerWithHandler(t, handler)

	defer server.Close()

	_, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",

		HTTPClient: client,
		OnRenewFunc: func(token string, expires time.Time) error {
			mu.Lock()
			defer mu.Unlock()

			savedToken = token
			savedExpires = expires

			return nil
		},
	})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	mu.Lock()
	if loginCallCount != 1 {
		mu.Unlock()
		t.Fatalf("expected 1 HTTP call for login, got %d", loginCallCount)
	}
	mu.Unlock()

	if savedToken != newToken {
		t.Fatalf("expected saved token to match generated JWT")
	}

	if !savedExpires.After(time.Now().Add(-time.Hour)) {
		t.Fatal("expected savedExpires to be recent")
	}
}

func TestApiGet_uses_httpclient(t *testing.T) {
	t.Parallel()

	mu := &sync.Mutex{}

	var called bool

	var apiPath string

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		called = true
		apiPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{{"id": "abc123", "name": "demo"}},
		})
		if err != nil {
			t.Fail()
		}
	}

	server, httpClient := testServerWithHandler(t, handler)

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWT(time.Now().Add(24 * time.Hour)),

		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Fatal("expected Clients() to call the HTTP client")
	}

	if apiPath != "/api/v2/clients" {
		t.Fatalf("expected path /api/v2/clients, got %s", apiPath)
	}
}

func TestClients_uses_httpclient(t *testing.T) {
	t.Parallel()

	mu := &sync.Mutex{}

	var called bool

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		called = true

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{{"id": "abc123", "name": "demo"}},
		})
		if err != nil {
			t.Fail()
		}
	}

	server, httpClient := testServerWithHandler(t, handler)

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWT(time.Now().Add(24 * time.Hour)),

		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Fatal("expected Clients() to call the HTTP client")
	}
}

func TestApiCall_uses_httpclient(t *testing.T) {
	t.Parallel()

	mu := &sync.Mutex{}

	var called bool

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		called = true

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`[{"id":"abc123","data":{"fullName":"Test"},"doc_id":[1]}]`))
		if err != nil {
			t.Fail()
		}
	}

	server, httpClient := testServerWithHandler(t, handler)

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWT(time.Now().Add(24 * time.Hour)),

		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if !called {
		t.Fatal("expected Users() to call the HTTP client")
	}
}

func TestCheckExpired_uses_httpclient(t *testing.T) {
	t.Parallel()

	mu := &sync.Mutex{}

	var renewCallCount int

	handler := func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "token/refresh") {
			renewCallCount++

			w.Header().Set("Content-Type", "application/json")

			newJwt := genJWT(time.Now().Add(24 * time.Hour))

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":    "success",
				"tenant_id": 42,
				"token":     newJwt,
			})
			if err != nil {
				t.Fail()
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{{"id": "abc123", "name": "demo"}},
		})
		if err != nil {
			t.Fail()
		}
	}

	server, httpClient := testServerWithHandler(t, handler)

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWT(time.Now().Add(1 * time.Minute)),

		HTTPClient: httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Clients() triggers checkExpired internally which triggers token renewal
	_, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if renewCallCount != 1 {
		t.Fatalf("expected 1 HTTP call for token refresh, got %d", renewCallCount)
	}
}
