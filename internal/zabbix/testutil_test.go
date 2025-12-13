package zabbix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockResponse represents a mock JSON-RPC response.
type mockResponse struct {
	Result any
	Error  *APIError
}

// newMockServer creates a test server that responds to JSON-RPC requests.
// The handler map keys are method names, values are the responses to return.
func newMockServer(t *testing.T, handlers map[string]mockResponse) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		handler, ok := handlers[req.Method]
		if !ok {
			t.Errorf("unexpected method: %s", req.Method)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		resp := Response{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   handler.Error,
		}

		if handler.Result != nil {
			resultBytes, err := json.Marshal(handler.Result)
			if err != nil {
				t.Errorf("failed to marshal result: %v", err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			resp.Result = resultBytes
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("failed to encode response: %v", err)
		}
	}))
}

// newTestClient creates a client pointing to the given test server.
func newTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	// Remove the /api_jsonrpc.php suffix that NewClient adds
	// since we're providing the full URL
	client := &Client{
		baseURL:    serverURL,
		httpClient: http.DefaultClient,
	}
	return client
}
