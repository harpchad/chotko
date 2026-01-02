package zabbix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://zabbix.example.com")

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	expectedURL := "https://zabbix.example.com/api_jsonrpc.php"
	if client.baseURL != expectedURL {
		t.Errorf("baseURL = %q, want %q", client.baseURL, expectedURL)
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	client := NewClient("https://zabbix.example.com", WithTimeout(5*time.Second))

	if client.httpClient.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want %v", client.httpClient.Timeout, 5*time.Second)
	}
}

func TestNewClient_WithInsecureSkipVerify(t *testing.T) {
	client := NewClient("https://zabbix.example.com", WithInsecureSkipVerify())

	transport, ok := client.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Transport is not *http.Transport")
	}

	if transport.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}

	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify = false, want true")
	}
}

func TestClient_SetToken(t *testing.T) {
	client := NewClient("https://zabbix.example.com")
	client.SetToken("test-token")

	if client.token != "test-token" {
		t.Errorf("token = %q, want %q", client.token, "test-token")
	}
}

func TestClient_Version(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"apiinfo.version": {Result: "7.0.0"},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)

	version, err := client.Version(context.Background())
	if err != nil {
		t.Fatalf("Version() error = %v", err)
	}

	if version != "7.0.0" {
		t.Errorf("Version() = %q, want %q", version, "7.0.0")
	}
}

func TestClient_Version_Error(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"apiinfo.version": {
			Error: &APIError{
				Code:    -32600,
				Message: "Invalid request",
				Data:    "Some error data",
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)

	_, err := client.Version(context.Background())
	if err == nil {
		t.Fatal("Version() expected error")
	}
}

func TestClient_IsConnected(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"apiinfo.version": {Result: "7.0.0"},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)

	if !client.IsConnected(context.Background()) {
		t.Error("IsConnected() = false, want true")
	}
}

func TestClient_IsConnected_Failure(t *testing.T) {
	// Server that doesn't respond properly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	if client.IsConnected(context.Background()) {
		t.Error("IsConnected() = true, want false")
	}
}

func TestClient_Login(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"user.login": {Result: "session-token-123"},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)

	err := client.Login(context.Background(), "admin", "password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if client.token != "session-token-123" {
		t.Errorf("token = %q, want %q", client.token, "session-token-123")
	}
}

func TestClient_Login_InvalidCredentials(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"user.login": {
			Error: &APIError{
				Code:    -32602,
				Message: "Invalid params",
				Data:    "Login name or password is incorrect.",
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)

	err := client.Login(context.Background(), "admin", "wrong-password")
	if err == nil {
		t.Fatal("Login() expected error for invalid credentials")
	}
}

func TestClient_Logout(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"user.logout": {Result: true},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "session-token"

	err := client.Logout(context.Background())
	if err != nil {
		t.Fatalf("Logout() error = %v", err)
	}

	if client.token != "" {
		t.Errorf("token = %q, want empty string after logout", client.token)
	}
}

func TestClient_Logout_NoToken(t *testing.T) {
	client := NewClient("https://zabbix.example.com")

	// Should not error when there's no token
	err := client.Logout(context.Background())
	if err != nil {
		t.Errorf("Logout() with no token error = %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError *APIError
		contains []string
	}{
		{
			name: "with data",
			apiError: &APIError{
				Code:    -32600,
				Message: "Invalid request",
				Data:    "Additional details",
			},
			contains: []string{"-32600", "Invalid request", "Additional details"},
		},
		{
			name: "without data",
			apiError: &APIError{
				Code:    -32600,
				Message: "Invalid request",
			},
			contains: []string{"-32600", "Invalid request"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.apiError.Error()
			for _, want := range tt.contains {
				if !containsStr(errMsg, want) {
					t.Errorf("Error() = %q, want to contain %q", errMsg, want)
				}
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Version(ctx)
	if err == nil {
		t.Error("Version() expected error for canceled context")
	}
}

func TestClient_NetworkError(t *testing.T) {
	// Client pointing to non-existent server
	client := NewClient("http://localhost:1") // Port 1 is typically not used

	_, err := client.Version(context.Background())
	if err == nil {
		t.Error("Version() expected error for network failure")
	}
}
