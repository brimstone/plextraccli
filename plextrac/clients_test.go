// Copyright (c) 2026 Matt Robinson brimstone@the.narro.ws

package plextrac_test

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/brimstone/plextraccli/plextrac"
)

const clientsEndpoint = "/api/v2/clients"
const clientDetails = "/api/v1/client/123"

// Helper function to check if a slice contains all elements of another slice (order doesn't matter).
func containsStringSlice(a, b []string) bool {
	if len(b) > len(a) {
		return false
	}

	for _, item := range b {
		if slices.Contains(a, item) {
			return true
		}
	}

	return false
}

// testServerWithHandlerClients creates a test server with the given handler function.
func testServerWithHandlerClients(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *http.Client) {
	t.Helper()

	server := httptest.NewTLSServer(handler)

	return server, &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test server only
		},
	}
}

// genJWTClients generates a fake JWT token for testing.
func genJWTClients(exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload := map[string]any{"exp": exp.Unix()}

	pBytes, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	encoded := strings.TrimRight(base64.RawURLEncoding.EncodeToString(pBytes), "=")

	return header + "." + encoded + ".fakesig"
}

// createTestClientsResponse creates a standard test response with client data.
func createTestClientsResponse() map[string]any {
	return map[string]any{
		"status": "success",
		"data": []map[string]any{
			{
				"client_id":   123,
				"name":        "Test Client",
				"description": "A test client",
				"poc":         "John Doe",
				"poc_email":   "john@example.com",
				"tags":        []string{"tag1", "tag2", "tag3"},
			},
			{
				"client_id":   456,
				"name":        "Another Client",
				"description": "Another test client",
				"poc":         "",
				"poc_email":   "",
				"tags":        []string{"tag3"},
			},
		},
	}
}

// createSingleClientResponse creates a test response with a single client.
func createSingleClientResponse() map[string]any {
	return map[string]any{
		"status": "success",
		"data": []map[string]any{
			{
				"client_id":   123,
				"name":        "Test Client",
				"description": "A test client",
				"poc":         "John Doe",
				"poc_email":   "john@example.com",
				"tags":        []string{"tag4", "tag5"},
			},
		},
	}
}

// createMultipleMatchResponse creates a test response with multiple matching clients.
func createMultipleMatchResponse() map[string]any {
	return map[string]any{
		"status": "success",
		"data": []map[string]any{
			{
				"client_id":   123,
				"name":        "Test Client",
				"description": "A test client",
				"poc":         "John Doe",
				"poc_email":   "john@example.com",
				"tags":        []string{"tag6", "tag7"},
			},
			{
				"client_id":   456,
				"name":        "Test Another",
				"description": "Another test client",
				"poc":         "",
				"poc_email":   "",
				"tags":        []string{"tag8"},
			},
		},
	}
}

// createHandler creates an HTTP handler function for client endpoint tests.
func createHandler(t *testing.T, called *bool, responseFunc func() map[string]any) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == clientsEndpoint && r.Method == http.MethodPost {
			*called = true

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(responseFunc())
			if err != nil {
				t.Fatalf("Failed to encode response: %v", err)
			}

			return
		}

		// For token refresh endpoint
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "token/refresh") {
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":    "success",
				"tenant_id": 42,
				"token":     genJWTClients(time.Now().Add(24 * time.Hour)),
			})
			if err != nil {
				t.Fatalf("Failed to encode token refresh response: %v", err)
			}

			return
		}

		// Default response for other endpoints
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{},
		})
		if err != nil {
			t.Fatalf("Failed to encode default response: %v", err)
		}
	}
}

func TestClients_Basic(t *testing.T) {
	t.Parallel()

	var called bool

	handler := createHandler(t, &called, createTestClientsResponse)

	server, httpClient := testServerWithHandlerClients(t, handler)
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	clients, err := ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if !called {
		t.Fatal("expected Clients() to call the HTTP client")
	}

	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}

	// Check first client
	if clients[0].ID != 123 {
		t.Fatalf("expected client ID 123, got %d", clients[0].ID)
	}

	if clients[0].Name != "Test Client" {
		t.Fatalf("expected client name 'Test Client', got '%s'", clients[0].Name)
	}

	if clients[0].Description != "A test client" {
		t.Fatalf("expected client description 'A test client', got '%s'", clients[0].Description)
	}

	if clients[0].POC != "John Doe" {
		t.Fatalf("expected client POC 'John Doe', got '%s'", clients[0].POC)
	}

	if clients[0].POCEmail != "john@example.com" {
		t.Fatalf("expected client POCEmail 'john@example.com', got '%s'", clients[0].POCEmail)
	}

	if !containsStringSlice(clients[0].Tags(), []string{"tag1", "tag2"}) {
		t.Fatalf("expected client tags ['tag1', 'tag2'], got %v", clients[0].Tags())
	}

	// Check second client
	if clients[1].ID != 456 {
		t.Fatalf("expected client ID 456, got %d", clients[1].ID)
	}

	if clients[1].Name != "Another Client" {
		t.Fatalf("expected client name 'Another Client', got '%s'", clients[1].Name)
	}

	if clients[1].Description != "Another test client" {
		t.Fatalf("expected client description 'Another test client', got '%s'", clients[1].Description)
	}

	if clients[1].POC != "" {
		t.Fatalf("expected client POC '', got '%s'", clients[1].POC)
	}

	if clients[1].POCEmail != "" {
		t.Fatalf("expected client POCEmail '', got '%s'", clients[1].POCEmail)
	}

	if !containsStringSlice(clients[1].Tags(), []string{"tag3"}) {
		t.Fatalf("expected client tags ['tag3'], got %v", clients[1].Tags())
	}
}

func TestClientByPartial_SingleMatch(t *testing.T) {
	t.Parallel()

	var called bool

	handler := createHandler(t, &called, createTestClientsResponse)

	server, httpClient := testServerWithHandlerClients(t, handler)
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	client, err := ua.ClientByPartial("Test")
	if err != nil {
		t.Fatalf("ClientByPartial() returned error: %v", err)
	}

	if !called {
		t.Fatal("expected ClientByPartial() to call the HTTP client")
	}

	if client == nil {
		t.Fatal("expected client to not be nil")
	}

	if client.ID != 123 {
		t.Fatalf("expected client ID 123, got %d", client.ID)
	}

	if client.Name != "Test Client" {
		t.Fatalf("expected client name 'Test Client', got '%s'", client.Name)
	}
}

func TestClientByPartial_NoMatch(t *testing.T) {
	t.Parallel()

	var called bool

	handler := createHandler(t, &called, createSingleClientResponse)

	server, httpClient := testServerWithHandlerClients(t, handler)
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	client, err := ua.ClientByPartial("NonExistent")
	if err == nil {
		t.Fatalf("expected error for non-existent client, got nil")
	}

	if client != nil {
		t.Fatalf("expected nil client for non-existent match, got %v", client)
	}

	if !called {
		t.Fatal("expected ClientByPartial() to call the HTTP client")
	}
}

func TestClientByPartial_MultipleMatches(t *testing.T) {
	t.Parallel()

	var called bool

	handler := createHandler(t, &called, createMultipleMatchResponse)

	server, httpClient := testServerWithHandlerClients(t, handler)
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	client, err := ua.ClientByPartial("Test")
	if err == nil {
		t.Fatalf("expected error for multiple matching clients, got nil")
	}

	if client != nil {
		t.Fatalf("expected nil client for multiple matches, got %v", client)
	}

	if !called {
		t.Fatal("expected ClientByPartial() to call the HTTP client")
	}
}
func clientHandler(t *testing.T, clientsCalled *bool, ensureFullCalled *bool, updateCalled *bool, updatedTags *[]string) func(http.ResponseWriter, *http.Request) {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		// Handle POST to /api/v2/clients (for ua.Clients())
		if r.URL.Path == clientsEndpoint && r.Method == http.MethodPost {
			*clientsCalled = true

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(createTestClientsResponse())
			if err != nil {
				t.Fatalf("Failed to encode response: %v", err)
			}

			return
		}

		// Handle GET to /api/v1/client/123 (for EnsureFull)
		if r.URL.Path == clientDetails && r.Method == http.MethodGet {
			*ensureFullCalled = true

			// Return some raw client data (minimal for the test)
			raw := map[string]any{
				"client_id":   123,
				"cuid":        "somecuid",
				"doc_type":    "client",
				"licenseKeys": []string{},
				"logo":        "",
				"tenant_id":   1,
				"users":       map[string]any{},
				// Include the existing tags so we can see them being appended
				"tags": []string{"tag1", "tag2"},
			}

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(raw)
			if err != nil {
				t.Fatalf("Failed to encode raw client response: %v", err)
			}

			return
		}

		// Handle PUT to /api/v1/client/123 (for update)
		if r.URL.Path == clientDetails && r.Method == http.MethodPut {
			*updateCalled = true

			// Decode the request body to see what tags were sent
			var raw map[string]any

			err := json.NewDecoder(r.Body).Decode(&raw)
			if err != nil {
				t.Fatalf("Failed to decode update request body: %v", err)
			}

			if tags, ok := raw["tags"].([]any); ok {
				for _, t := range tags {
					if str, ok := t.(string); ok {
						*updatedTags = append(*updatedTags, str)
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")

			err = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
			})
			if err != nil {
				t.Fatalf("Failed to encode update response: %v", err)
			}

			return
		}

		// For token refresh endpoint
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "token/refresh") {
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":    "success",
				"tenant_id": 42,
				"token":     genJWTClients(time.Now().Add(24 * time.Hour)),
			})
			if err != nil {
				t.Fatalf("Failed to encode token refresh response: %v", err)
			}

			return
		}

		// Default response for other endpoints (should not happen in test)
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{},
		})
		if err != nil {
			t.Fatalf("Failed to encode default response: %v", err)
		}
	}
}

// TestClient_AddTags tests the AddTags method on a Client.
func TestClient_AddTags(t *testing.T) {
	t.Parallel()

	var (
		clientsCalled    bool
		ensureFullCalled bool
		updateCalled     bool
		updatedTags      []string
	)

	server, httpClient := testServerWithHandler(t, clientHandler(t, &clientsCalled, &ensureFullCalled, &updateCalled, &updatedTags))
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Get the client to test on
	var clients []*plextrac.Client

	clients, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if len(clients) == 0 {
		t.Fatal("expected at least one client")
	}

	client := clients[0]

	// Call AddTags with a new tag
	newTag := "newtag"

	warnings, err := client.AddTags([]string{newTag})
	if err != nil {
		t.Fatalf("AddTags returned error: %v", err)
	}
	// We expect no warnings in this test
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}

	// Check that the client's tags were updated in memory
	if !containsStringSlice(client.Tags(), []string{"tag1", "tag2", newTag}) {
		t.Fatalf("expected tags to contain tag1, tag2, newtag, got %v", client.Tags())
	}

	// Verify that the HTTP calls were made as expected
	if !clientsCalled {
		t.Fatal("expected clients endpoint to be called")
	}

	if !ensureFullCalled {
		t.Fatal("expected EnsureFull to be called (GET to v1/client/123)")
	}

	if !updateCalled {
		t.Fatal("expected update to be called (PUT to v1/client/123)")
	}
	// Check that the updated tags include the new tag along with the existing ones
	if !containsStringSlice(updatedTags, []string{"tag1", "tag2", newTag}) {
		t.Fatalf("expected updated tags to contain tag1, tag2, newtag, got %v", updatedTags)
	}
}

// TestClient_RemoveTags tests the RemoveTags method on a Client.
func TestClient_RemoveTags(t *testing.T) {
	t.Parallel()

	var (
		clientsCalled    bool
		ensureFullCalled bool
		updateCalled     bool
		updatedTags      []string
	)

	server, httpClient := testServerWithHandler(t, clientHandler(t, &clientsCalled, &ensureFullCalled, &updateCalled, &updatedTags))

	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Get the client to test on
	var clients []*plextrac.Client

	clients, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if len(clients) == 0 {
		t.Fatal("expected at least one client")
	}

	client := clients[0]

	fmt.Printf("test client tags: %#v\n", client.Tags())

	// Call RemoveTags with a tag to remove
	tagToRemove := "tag2"

	warnings, err := client.RemoveTags([]string{tagToRemove})
	if err != nil {
		t.Fatalf("RemoveTags returned error: %v", err)
	}
	// We expect no warnings in this test
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}

	// Check that the client's tags were updated in memory (tag2 should be removed)
	if !containsStringSlice(client.Tags(), []string{"tag1", "tag3"}) {
		t.Fatalf("expected tags to contain tag1 and tag3 (tag2 removed), got %v", client.Tags())
	}

	// Verify that the HTTP calls were made as expected
	if !clientsCalled {
		t.Fatal("expected clients endpoint to be called")
	}

	if !ensureFullCalled {
		t.Fatal("expected EnsureFull to be called (GET to v1/client/123)")
	}

	if !updateCalled {
		t.Fatal("expected update to be called (PUT to v1/client/123)")
	}
	// Check that the updated tags have the tag removed
	if !containsStringSlice(updatedTags, []string{"tag1", "tag3"}) {
		t.Fatalf("expected updated tags to contain tag1 and tag3 (tag2 removed), got %v", updatedTags)
	}
}

// TestClient_SetTags tests the SetTags method on a Client.
func TestClient_SetTags(t *testing.T) {
	t.Parallel()

	var (
		clientsCalled    bool
		ensureFullCalled bool
		updateCalled     bool
		updatedTags      []string
	)

	// Handler for the specific test
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Handle POST to /api/v2/clients (for ua.Clients())
		if r.URL.Path == clientsEndpoint && r.Method == http.MethodPost {
			clientsCalled = true

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(createTestClientsResponse())
			if err != nil {
				t.Fatalf("Failed to encode response: %v", err)
			}

			return
		}

		// Handle GET to /api/v1/client/123 (for EnsureFull)
		if r.URL.Path == "/api/v1/client/123" && r.Method == http.MethodGet {
			ensureFullCalled = true

			// Return some raw client data (minimal for the test)
			raw := map[string]any{
				"client_id":   123,
				"cuid":        "somecuid",
				"doc_type":    "client",
				"licenseKeys": []string{},
				"logo":        "",
				"tenant_id":   1,
				"users":       map[string]any{},
				// Include the existing tags so we can see them being replaced
				"tags": []string{"tag1", "tag2"},
			}

			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(raw)
			if err != nil {
				t.Fatalf("Failed to encode raw client response: %v", err)
			}

			return
		}

		// Handle PUT to /api/v1/client/123 (for update)
		if r.URL.Path == "/api/v1/client/123" && r.Method == http.MethodPut {
			updateCalled = true

			// Decode the request body to see what tags were sent
			var raw map[string]any

			err := json.NewDecoder(r.Body).Decode(&raw)
			if err != nil {
				t.Fatalf("Failed to decode update request body: %v", err)
			}

			if tags, ok := raw["tags"].([]any); ok {
				for _, t := range tags {
					if str, ok := t.(string); ok {
						updatedTags = append(updatedTags, str)
					}
				}
			}

			w.Header().Set("Content-Type", "application/json")

			err = json.NewEncoder(w).Encode(map[string]any{
				"status": "success",
			})
			if err != nil {
				t.Fatalf("Failed to encode update response: %v", err)
			}

			return
		}

		// For token refresh endpoint
		if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "token/refresh") {
			w.Header().Set("Content-Type", "application/json")

			err := json.NewEncoder(w).Encode(map[string]any{
				"status":    "success",
				"tenant_id": 42,
				"token":     genJWTClients(time.Now().Add(24 * time.Hour)),
			})
			if err != nil {
				t.Fatalf("Failed to encode token refresh response: %v", err)
			}

			return
		}

		// Default response for other endpoints (should not happen in test)
		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(map[string]any{
			"status": "success",
			"data":   []map[string]any{},
		})
		if err != nil {
			t.Fatalf("Failed to encode default response: %v", err)
		}
	}

	server, httpClient := testServerWithHandler(t, handler)
	defer server.Close()

	ua, _, err := plextrac.New(plextrac.NewOptions{
		InstanceURL: server.Listener.Addr().String(),
		AuthToken:   genJWTClients(time.Now().Add(24 * time.Hour)),
		HTTPClient:  httpClient,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	// Get the client to test on
	var clients []*plextrac.Client

	clients, err = ua.Clients()
	if err != nil {
		t.Fatalf("Clients() returned error: %v", err)
	}

	if len(clients) == 0 {
		t.Fatal("expected at least one client")
	}

	client := clients[0]

	// Call SetTags with new tags (replacing existing ones)
	newTags := []string{"newtag1", "newtag2"}

	warnings, err := client.SetTags(newTags)
	if err != nil {
		t.Fatalf("SetTags returned error: %v", err)
	}
	// We expect no warnings in this test
	if len(warnings) != 0 {
		t.Fatalf("expected no warnings, got %v", warnings)
	}

	// Check that the client's tags were updated in memory (should be exactly newTags)
	if !slices.Equal(client.Tags(), newTags) {
		t.Fatalf("expected tags to be %v, got %v", newTags, client.Tags())
	}

	// Verify that the HTTP calls were made as expected
	if !clientsCalled {
		t.Fatal("expected clients endpoint to be called")
	}

	if !ensureFullCalled {
		t.Fatal("expected EnsureFull to be called (GET to v1/client/123)")
	}

	if !updateCalled {
		t.Fatal("expected update to be called (PUT to v1/client/123)")
	}
	// Check that the updated tags are exactly the new tags
	if !slices.Equal(updatedTags, newTags) {
		t.Fatalf("expected updated tags to be %v, got %v", newTags, updatedTags)
	}
}
