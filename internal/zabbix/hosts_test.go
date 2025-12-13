package zabbix

import (
	"context"
	"testing"
)

func TestDefaultHostGetParams(t *testing.T) {
	params := DefaultHostGetParams()

	// Check that output is set
	if params.Output == nil {
		t.Error("Output not set in default params")
	}

	// Check selectInterfaces is set
	if params.SelectInterfaces == nil {
		t.Error("SelectInterfaces not set in default params")
	}

	// Check MonitoredHosts is true
	if !params.MonitoredHosts {
		t.Error("MonitoredHosts should be true")
	}
}

func TestClient_GetHosts(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Result: []Host{
				{HostID: "1", Host: "host1", Name: "Host 1"},
				{HostID: "2", Host: "host2", Name: "Host 2"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	hosts, err := client.GetHosts(context.Background(), DefaultHostGetParams())
	if err != nil {
		t.Fatalf("GetHosts() error = %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("len(hosts) = %d, want 2", len(hosts))
	}

	if hosts[0].Name != "Host 1" {
		t.Errorf("hosts[0].Name = %q, want %q", hosts[0].Name, "Host 1")
	}
}

func TestClient_GetHosts_Empty(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {Result: []Host{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	hosts, err := client.GetHosts(context.Background(), DefaultHostGetParams())
	if err != nil {
		t.Fatalf("GetHosts() error = %v", err)
	}

	if len(hosts) != 0 {
		t.Errorf("len(hosts) = %d, want 0", len(hosts))
	}
}

func TestClient_GetHosts_Error(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Error: &APIError{
				Code:    -32602,
				Message: "Invalid params",
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	_, err := client.GetHosts(context.Background(), DefaultHostGetParams())
	if err == nil {
		t.Fatal("GetHosts() expected error")
	}
}

func TestClient_GetAllHosts(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Result: []Host{
				{HostID: "1", Host: "host1", Name: "Host 1", Status: "0"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	hosts, err := client.GetAllHosts(context.Background())
	if err != nil {
		t.Fatalf("GetAllHosts() error = %v", err)
	}

	if len(hosts) != 1 {
		t.Errorf("len(hosts) = %d, want 1", len(hosts))
	}
}

func TestClient_GetHostCounts(t *testing.T) {
	// Test the critical availability mapping:
	// 0 = Unknown
	// 1 = Available (OK)
	// 2 = Unavailable (Problem)
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Result: []Host{
				// Available hosts (active_available = "1")
				{HostID: "1", Status: "0", MaintenanceStatus: "0", ActiveAvailable: "1"},
				{HostID: "2", Status: "0", MaintenanceStatus: "0", ActiveAvailable: "1"},
				// Unavailable host (active_available = "2")
				{HostID: "3", Status: "0", MaintenanceStatus: "0", ActiveAvailable: "2"},
				// Unknown host (active_available = "0")
				{HostID: "4", Status: "0", MaintenanceStatus: "0", ActiveAvailable: "0"},
				// Maintenance host
				{HostID: "5", Status: "0", MaintenanceStatus: "1", ActiveAvailable: "1"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	counts, err := client.GetHostCounts(context.Background())
	if err != nil {
		t.Fatalf("GetHostCounts() error = %v", err)
	}

	// Verify counts based on availability mapping
	// OK = Available (active_available = "1") and not in maintenance = 2
	if counts.OK != 2 {
		t.Errorf("counts.OK = %d, want 2", counts.OK)
	}

	// Problem = Unavailable (active_available = "2") = 1
	if counts.Problem != 1 {
		t.Errorf("counts.Problem = %d, want 1", counts.Problem)
	}

	// Unknown = active_available = "0" = 1
	if counts.Unknown != 1 {
		t.Errorf("counts.Unknown = %d, want 1", counts.Unknown)
	}

	// Maintenance = 1
	if counts.Maintenance != 1 {
		t.Errorf("counts.Maintenance = %d, want 1", counts.Maintenance)
	}

	// Total = 5
	if counts.Total != 5 {
		t.Errorf("counts.Total = %d, want 5", counts.Total)
	}
}

func TestClient_GetHostCounts_Empty(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {Result: []Host{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	counts, err := client.GetHostCounts(context.Background())
	if err != nil {
		t.Fatalf("GetHostCounts() error = %v", err)
	}

	if counts.Total != 0 {
		t.Errorf("counts.Total = %d, want 0", counts.Total)
	}
}

func TestClient_GetHost(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Result: []Host{
				{HostID: "123", Host: "test-host", Name: "Test Host"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	host, err := client.GetHost(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetHost() error = %v", err)
	}

	if host == nil {
		t.Fatal("GetHost() returned nil")
	}

	if host.HostID != "123" {
		t.Errorf("host.HostID = %q, want %q", host.HostID, "123")
	}
}

func TestClient_GetHost_NotFound(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {Result: []Host{}},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	_, err := client.GetHost(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("GetHost() expected error for non-existent host")
	}
}

func TestClient_SearchHosts(t *testing.T) {
	server := newMockServer(t, map[string]mockResponse{
		"host.get": {
			Result: []Host{
				{HostID: "1", Host: "web-server-1", Name: "Web Server 1"},
				{HostID: "2", Host: "web-server-2", Name: "Web Server 2"},
			},
		},
	})
	defer server.Close()

	client := newTestClient(t, server.URL)
	client.token = "test-token"

	hosts, err := client.SearchHosts(context.Background(), "web")
	if err != nil {
		t.Fatalf("SearchHosts() error = %v", err)
	}

	if len(hosts) != 2 {
		t.Errorf("len(hosts) = %d, want 2", len(hosts))
	}
}
