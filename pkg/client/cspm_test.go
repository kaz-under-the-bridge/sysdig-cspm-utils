package client

import (
	"testing"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/internal/testutil"
)

func TestNewCSPMClient(t *testing.T) {
	client := NewCSPMClient("https://test.example.com", "test-token")

	if client == nil {
		t.Fatal("NewCSPMClient returned nil")
	}

	if client.Client == nil {
		t.Error("CSPMClient.Client is nil")
	}
}

func TestGetComplianceRequirements(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	tests := []struct {
		name      string
		filter    string
		wantError bool
		checkData bool
	}{
		{
			name:      "CIS AWS フィルター",
			filter:    `pass = "false" and policy.name in ("CIS Amazon Web Services Foundations Benchmark v3.0.0") and zone.name in ("Entire Infrastructure")`,
			wantError: false,
			checkData: true,
		},
		{
			name:      "CIS GCP フィルター",
			filter:    `pass = "false" and policy.name in ("CIS Google Cloud Platform Foundation Benchmark v2.0.0") and zone.name in ("Entire Infrastructure")`,
			wantError: false,
			checkData: true,
		},
		{
			name:      "空フィルター",
			filter:    "",
			wantError: false,
			checkData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GetComplianceRequirements(tt.filter)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Error("Response is nil")
				return
			}

			if tt.checkData {
				if response.Data == nil {
					t.Error("Response.Data is nil")
				}
			}
		})
	}
}

func TestGetComplianceRequirementsPaginated(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	tests := []struct {
		name       string
		filter     string
		pageNumber int
		pageSize   int
		wantError  bool
	}{
		{
			name:       "ページ1",
			filter:     "",
			pageNumber: 1,
			pageSize:   2,
			wantError:  false,
		},
		{
			name:       "ページ2",
			filter:     "",
			pageNumber: 2,
			pageSize:   2,
			wantError:  false,
		},
		{
			name:       "デフォルトページング",
			filter:     "",
			pageNumber: 0,
			pageSize:   0,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GetComplianceRequirementsPaginated(tt.filter, tt.pageNumber, tt.pageSize)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if response == nil {
				t.Error("Response is nil")
			}
		})
	}
}

func TestGetAllComplianceRequirements(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	// フィルター付きでテスト（モックサーバーが対応しているフィルター）
	response, err := client.GetAllComplianceRequirements(`pass = "false" and policy.name in ("CIS Amazon Web Services Foundations Benchmark v3.0.0")`, 2, 1, 0)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if response == nil {
		t.Error("Response is nil")
		return
	}

	if len(response.Data) == 0 {
		t.Error("Expected data but got empty array")
	}
}
