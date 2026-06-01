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

// genUserJWT creates a JWT with the tenantId claim set to 42, required for tenant-aware API calls.
func genUserJWT(t *testing.T, exp time.Time) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := map[string]any{"exp": exp.Unix(), "tenantId": 42}

	pBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	encoded := strings.TrimRight(base64.RawURLEncoding.EncodeToString(pBytes), "=")

	return header + "." + encoded + ".fakesig"
}

// userMock represents the configurable parts of a user-reset mock handler.
type userMock struct {
	userID       string
	userFullName string
	userEmail    string
	userIdInt    int64
	status       string
	message      string
}

// makeUserHandler returns an http.HandlerFunc that handles user listing and
// password-reset endpoints with the given mock configuration.
func makeUserHandler(mock userMock) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":  mock.status,
				"message": mock.message,
			})
			if err != nil {
				panic(err)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     mock.userID,
				"status": "success",
				"doc_id": []any{mock.userIdInt},
				"data":   map[string]any{"doc_type": "user", "fullName": mock.userFullName, "email": mock.userEmail, "disabled": false, "user_id": mock.userIdInt, "createdAt": time.Now().UnixMilli(), "lastLogin": time.Now().UnixMilli()},
			},
		})
		if err != nil {
			panic(err)
		}
	}
}

// makeUserServer creates a test server and returns (httpClient, server).
// The server handles login at /api/v1/authenticate and delegates all other paths
// to the provided userHandler, which must be safe to call without lock.
func makeUserServer(t *testing.T, userHandler http.HandlerFunc) (*http.Client, *httptest.Server) {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/authenticate" {
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":      "success",
				"tenant_id":   42,
				"token":       genUserJWT(t, time.Now().Add(24*time.Hour)),
				"mfa_enabled": false,
				"cookie":      "session=test",
				"statusCode":  200,
				"error":       "",
				"message":     "",
			})
			if err != nil {
				t.Fatal(err)
			}

			return
		}

		userHandler(w, r)
	})

	server := httptest.NewTLSServer(handler)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test server only
		},
	}

	return client, server
}

func TestUsers_returns_users_from_api(t *testing.T) {
	t.Parallel()

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     "1",
				"status": "success",
				"doc_id": []any{int64(1)},
				"data":   map[string]any{"doc_type": "user", "fullName": "Jane Doe", "email": "jane@example.com", "disabled": false, "user_id": int64(1), "createdAt": time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(), "lastLogin": time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC).UnixMilli()},
			},
			map[string]any{
				"id":     "2",
				"status": "success",
				"doc_id": []any{int64(2)},
				"data":   map[string]any{"doc_type": "user", "fullName": "John Smith", "email": "john@example.com", "disabled": true, "user_id": int64(2), "createdAt": time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC).UnixMilli(), "lastLogin": time.Date(2024, 2, 28, 8, 30, 0, 0, time.UTC).UnixMilli()},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}

	if users[0].ID == 0 {
		t.Error("expected first user ID to be non-zero")
	}

	if users[0].Name != "Jane Doe" {
		t.Errorf("expected first user name 'Jane Doe', got %q", users[0].Name)
	}

	if users[0].Email != "jane@example.com" {
		t.Errorf("expected first user email 'jane@example.com', got %q", users[0].Email)
	}

	if !users[0].Enabled {
		t.Error("expected first user to be enabled")
	}

	if users[1].ID == 0 {
		t.Error("expected second user ID to be non-zero")
	}

	if users[1].Name != "John Smith" {
		t.Errorf("expected second user name 'John Smith', got %q", users[1].Name)
	}

	if users[1].Enabled {
		t.Error("expected second user to be disabled")
	}
}

func TestUsers_empty_list(t *testing.T) {
	t.Parallel()

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	if len(users) != 0 {
		t.Fatalf("expected 0 users, got %d", len(users))
	}
}

func TestUsers_handles_api_error(t *testing.T) {
	t.Parallel()

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status":     "error",
			"message":    "Internal Server Error",
			"statusCode": 500,
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	_, err = ua.Users()
	if err == nil {
		t.Fatal("expected Users() to return an error")
	}
}

func TestUsers_sends_tenant_id_in_path(t *testing.T) {
	t.Parallel()

	var (
		gotPath string
		mu      sync.Mutex
	)

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		gotPath = r.URL.Path

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     "1",
				"status": "success",
				"doc_id": []any{int64(1)},
				"data":   map[string]any{"doc_type": "user", "fullName": "Jane Doe", "email": "jane@example.com", "disabled": false, "user_id": int64(1), "createdAt": time.Now().UnixMilli(), "lastLogin": time.Now().UnixMilli()},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
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

	if !strings.Contains(gotPath, "tenant") || !strings.Contains(gotPath, "user/list") {
		t.Errorf("expected path to contain 'tenant' and 'user/list', got %s", gotPath)
	}
}

func TestUsers_parses_timestamps_with_millis(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 3, 15, 10, 30, 0, 0, time.UTC)

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		milliTime := fixedTime.UnixMilli()

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     "5",
				"status": "success",
				"doc_id": []any{int64(5)},
				"data":   map[string]any{"doc_type": "user", "fullName": "Timing Test", "email": "timing@demo.example.com", "disabled": false, "user_id": int64(5), "createdAt": milliTime, "lastLogin": milliTime},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	u := users[0]
	if u.CreatedAt.Unix() != fixedTime.Unix() {
		t.Errorf("expected createdAt %d, got %d", fixedTime.Unix(), u.CreatedAt.Unix())
	}

	if u.LastLogin.Unix() != fixedTime.Unix() {
		t.Errorf("expected lastLogin %d, got %d", fixedTime.Unix(), u.LastLogin.Unix())
	}
}

func TestUser_String(t *testing.T) {
	t.Parallel()

	user := &plextrac.User{
		ID:    42,
		Name:  "Jane Doe",
		Email: "jane@example.com",
	}

	want := "Jane Doe <jane@example.com>"
	if got := user.String(); got != want {
		t.Errorf("User.String() = %q, want %q", got, want)
	}
}

func TestUser_Reset_returns_success(t *testing.T) {
	t.Parallel()

	var (
		resetCalled bool
	)

	mu := sync.Mutex{}

	helper := makeUserHandler(userMock{
		userID:       "42",
		userIdInt:    42,
		userFullName: "Reset User",
		userEmail:    "reset@demo.example.com",
		status:       "success",
		message:      "Password has been reset",
	})

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mu.Lock()
			resetCalled = true
			mu.Unlock()
			helper(w, r)

			return
		}

		helper(w, r)
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	errs, err := users[0].Reset()
	if err != nil {
		t.Fatalf("Reset() returned error: %v", err)
	}

	if errs != nil {
		t.Fatalf("Reset() returned errors: %v", errs)
	}

	mu.Lock()
	if !resetCalled {
		mu.Unlock()
		t.Fatal("expected Reset() to make a PUT request")
	}
	mu.Unlock()
}

func TestUser_Reset_handles_failure_response(t *testing.T) {
	t.Parallel()

	var (
		resetCalled bool
	)

	mu := sync.Mutex{}

	helper := makeUserHandler(userMock{
		userID:       "99",
		userIdInt:    99,
		userFullName: "Bad User",
		userEmail:    "bad@demo.example.com",
		status:       "error",
		message:      "Permission denied",
	})

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mu.Lock()
			resetCalled = true
			mu.Unlock()
			helper(w, r)

			return
		}

		helper(w, r)
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	_, err = users[0].Reset()
	if err == nil {
		t.Fatal("expected Reset() to return an error for failure response")
	}

	mu.Lock()
	if !resetCalled {
		mu.Unlock()
		t.Fatal("expected Reset() to make a PUT request")
	}
	mu.Unlock()
}

func TestUser_Reset_sends_correct_method_and_path(t *testing.T) {
	t.Parallel()

	var (
		mu                          sync.Mutex
		gotMethod, gotPath, gotBody string
	)

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut {
			gotMethod = r.Method
			gotPath = r.URL.Path

			body := make([]byte, r.ContentLength)
			if r.ContentLength > 0 {
				_, _ = r.Body.Read(body)
			}

			gotBody = string(body)

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":  "success",
				"message": "Password reset",
			})
			if err != nil {
				t.Fatal(err)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     "42",
				"status": "success",
				"doc_id": []any{int64(42)},
				"data":   map[string]any{"doc_type": "user", "fullName": "Reset User", "email": "reset@demo.example.com", "disabled": false, "user_id": int64(42), "createdAt": time.Now().UnixMilli(), "lastLogin": time.Now().UnixMilli()},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	errs, err := users[0].Reset()
	if err != nil {
		t.Fatalf("Reset() returned error: %v", err)
	}

	if errs != nil {
		t.Fatalf("Reset() returned errors: %v", errs)
	}

	mu.Lock()
	defer mu.Unlock()

	if gotMethod != "PUT" {
		t.Errorf("expected method PUT, got %s", gotMethod)
	}

	if gotPath != "/api/v1/tenant/42/user/resetpass" {
		t.Errorf("expected path /api/v1/tenant/42/user/resetpass, got %s", gotPath)
	}

	if gotBody == "" {
		t.Error("expected non-empty request body")
	}
}

func TestUser_Reset_sends_username_in_body(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	type request struct {
		Username string `json:"username"`
	}

	var gotRequest request

	client, server := makeUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		if r.Method == http.MethodPut {
			err := json.NewDecoder(r.Body).Decode(&gotRequest)
			if err != nil {
				t.Fatal(err)
			}

			w.Header().Set("Content-Type", "application/json")

			err = json.NewEncoder(w).Encode(map[string]any{
				"status":  "success",
				"message": "Password reset",
			})
			if err != nil {
				t.Fatal(err)
			}

			return
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode([]any{
			map[string]any{
				"id":     "42",
				"status": "success",
				"doc_id": []any{int64(42)},
				"data":   map[string]any{"doc_type": "user", "fullName": "Reset User", "email": "reset@demo.example.com", "disabled": false, "user_id": int64(42), "createdAt": time.Now().UnixMilli(), "lastLogin": time.Now().UnixMilli()},
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		Username:    "testuser",
		Password:    "testpass",
		HTTPClient:  client,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	users, err := ua.Users()
	if err != nil {
		t.Fatalf("Users() returned error: %v", err)
	}

	_, _ = users[0].Reset()

	mu.Lock()
	defer mu.Unlock()

	if gotRequest.Username != "reset@demo.example.com" {
		t.Errorf("expected username in body 'reset@demo.example.com', got %q", gotRequest.Username)
	}
}
