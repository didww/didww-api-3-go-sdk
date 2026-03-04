package didww

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// fixturesDir returns the absolute path to the testdata/fixtures directory.
func fixturesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata", "fixtures")
}

// loadFixture reads a fixture file from the testdata/fixtures directory.
func loadFixture(t *testing.T, path string) []byte {
	t.Helper()
	fullPath := filepath.Join(fixturesDir(), path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", path, err)
	}
	return data
}

// newTestServer creates an httptest.Server that serves fixture files based on the request path.
// The handler maps URL paths to fixture file paths.
func newTestServer(t *testing.T, routes map[string]testRoute) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			// Try without method prefix for simple GET routes
			route, ok = routes[r.URL.Path]
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	}))

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
	})

	return server, client
}

// testRoute defines a route for the test server.
type testRoute struct {
	status  int
	fixture string // path to fixture file relative to testdata/fixtures/
	body    string // raw body to return (used if fixture is empty)
}

// testServerWithClient wraps a test server with an associated client.
type testServerWithClient struct {
	server *http.Server
	client *Client
}

// newTestServerWithInspector creates a test server that calls an inspector function for each request.
func newTestServerWithInspector(t *testing.T, routes map[string]testRoute, inspector func(r *http.Request)) *testServerWithClient {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspector != nil {
			inspector(r)
		}

		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			route, ok = routes[r.URL.Path]
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	}))

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
	})

	return &testServerWithClient{client: client}
}

// newTestServerWithDynamicPatch creates a test server where PATCH requests are handled
// dynamically based on call count. Other methods use static routes.
func newTestServerWithDynamicPatch(t *testing.T, routes map[string]testRoute, inspector func(r *http.Request), patchRouter func(call int) testRoute) *testServerWithClient {
	t.Helper()
	var patchCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if inspector != nil {
			inspector(r)
		}

		var route testRoute
		var ok bool
		if r.Method == http.MethodPatch {
			patchCount++
			route = patchRouter(patchCount)
			ok = true
		} else {
			route, ok = routes[r.Method+" "+r.URL.Path]
			if !ok {
				route, ok = routes[r.URL.Path]
			}
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"errors":[{"title":"not found","status":"404"}]}`))
			return
		}

		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(route.status)
		if route.fixture != "" {
			data := loadFixture(t, route.fixture)
			w.Write(data)
		} else if route.body != "" {
			w.Write([]byte(route.body))
		}
	}))

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create test client: %v", err)
	}

	t.Cleanup(func() {
		server.Close()
	})

	return &testServerWithClient{client: client}
}

// assertRequestJSON compares the actual request body against a fixture file using semantic JSON comparison.
// Key order is ignored; both are normalized to map[string]any before comparison.
func assertRequestJSON(t *testing.T, actual []byte, fixturePath string) {
	t.Helper()
	expected := loadFixture(t, fixturePath)

	var actualObj any
	if err := json.Unmarshal(actual, &actualObj); err != nil {
		t.Fatalf("failed to parse actual request body: %v", err)
	}

	var expectedObj any
	if err := json.Unmarshal(expected, &expectedObj); err != nil {
		t.Fatalf("failed to parse expected fixture %s: %v", fixturePath, err)
	}

	if !reflect.DeepEqual(actualObj, expectedObj) {
		actualPretty, _ := json.MarshalIndent(actualObj, "", "  ")
		expectedPretty, _ := json.MarshalIndent(expectedObj, "", "  ")
		t.Errorf("request body mismatch for fixture %s\nGot:\n%s\nWant:\n%s", fixturePath, actualPretty, expectedPretty)
	}
}
