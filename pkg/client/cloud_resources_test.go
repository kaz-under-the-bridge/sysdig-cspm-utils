package client

import (
	"testing"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/internal/testutil"
)

func TestGetComplianceRequirementsWithControls(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	t.Run("get compliance requirements with controls - page 1", func(t *testing.T) {
		resp, err := client.GetComplianceRequirementsWithControls("", 1, 10)
		if err != nil {
			t.Fatalf("Failed to get compliance requirements: %v", err)
		}

		if len(resp.Data) == 0 {
			t.Error("Expected data, got empty")
		}

		// Check that controls are included
		if len(resp.Data[0].Controls) == 0 {
			t.Error("Expected controls in first requirement")
		}

		// Validate control structure
		ctrl := resp.Data[0].Controls[0]
		if ctrl.ID == "" {
			t.Error("Expected control ID")
		}
		if ctrl.ResourceAPIEndpoint == "" {
			t.Error("Expected resourceApiEndpoint")
		}
	})

	t.Run("get compliance requirements with controls - page 2", func(t *testing.T) {
		resp, err := client.GetComplianceRequirementsWithControls("", 2, 10)
		if err != nil {
			t.Fatalf("Failed to get compliance requirements: %v", err)
		}

		if len(resp.Data) == 0 {
			t.Error("Expected data, got empty")
		}
	})
}

func TestGetCloudResources(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	tests := []struct {
		name        string
		endpoint    string
		pageNumber  int
		pageSize    int
		expectError bool
		expectData  bool
	}{
		{
			name:        "control 16071 - page 1",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=16071",
			pageNumber:  1,
			pageSize:    50,
			expectError: false,
			expectData:  true,
		},
		{
			name:        "control 16071 - page 2",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=16071",
			pageNumber:  2,
			pageSize:    50,
			expectError: false,
			expectData:  true,
		},
		{
			name:        "control 16027 - S3 MFA Delete",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=16027",
			pageNumber:  1,
			pageSize:    50,
			expectError: false,
			expectData:  true,
		},
		{
			name:        "control 16026 - S3 Versioning",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=16026",
			pageNumber:  1,
			pageSize:    50,
			expectError: false,
			expectData:  true,
		},
		{
			name:        "control 16018 - IAM Policy",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=16018",
			pageNumber:  1,
			pageSize:    50,
			expectError: false,
			expectData:  true,
		},
		{
			name:        "unknown control - should return empty",
			endpoint:    "/api/cspm/v1/cloud/resources?controlId=99999",
			pageNumber:  1,
			pageSize:    50,
			expectError: false,
			expectData:  false,
		},
		{
			name:        "missing controlId - should error",
			endpoint:    "/api/cspm/v1/cloud/resources",
			pageNumber:  1,
			pageSize:    50,
			expectError: true,
			expectData:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.GetCloudResources(tt.endpoint, tt.pageNumber, tt.pageSize)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectData {
				if len(resp.Data) == 0 {
					t.Error("Expected data, got empty")
				}

				// Validate resource structure
				res := resp.Data[0]
				if res.Hash == "" {
					t.Error("Expected resource hash")
				}
				if res.Name == "" {
					t.Error("Expected resource name")
				}
				if res.Type == "" {
					t.Error("Expected resource type")
				}

				// Test GetAcceptanceStatus
				status := res.GetAcceptanceStatus()
				if status != "failed" && status != "passed" && status != "accepted" {
					t.Errorf("Invalid acceptance status: %s", status)
				}
			} else {
				if len(resp.Data) != 0 {
					t.Error("Expected empty data")
				}
			}
		})
	}
}

func TestCloudResourceAcceptanceStatus(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	// Get resources that have different acceptance statuses
	resp, err := client.GetCloudResources("/api/cspm/v1/cloud/resources?controlId=16027", 1, 50)
	if err != nil {
		t.Fatalf("Failed to get cloud resources: %v", err)
	}

	failedCount := 0
	passedCount := 0
	acceptedCount := 0

	for _, res := range resp.Data {
		status := res.GetAcceptanceStatus()
		switch status {
		case "failed":
			failedCount++
			if res.Passed {
				t.Errorf("Resource %s marked as failed but passed is true", res.Name)
			}
			if res.Acceptance != nil {
				t.Errorf("Resource %s marked as failed but has acceptance", res.Name)
			}
		case "passed":
			passedCount++
			if !res.Passed {
				t.Errorf("Resource %s marked as passed but passed is false", res.Name)
			}
			if res.Acceptance != nil {
				t.Errorf("Resource %s marked as passed but has acceptance", res.Name)
			}
		case "accepted":
			acceptedCount++
			if res.Acceptance == nil {
				t.Errorf("Resource %s marked as accepted but has no acceptance", res.Name)
			}
		}
	}

	t.Logf("Found %d failed, %d passed, %d accepted resources", failedCount, passedCount, acceptedCount)
}

func TestGetAllCloudResources(t *testing.T) {
	server := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer server.Close()

	client := NewCSPMClient(server.URL, "test-token")

	t.Run("get all resources for control with pagination", func(t *testing.T) {
		// This should fetch multiple pages
		resp, err := client.GetAllCloudResources("/api/cspm/v1/cloud/resources?controlId=16071", 2, 2, 0)
		if err != nil {
			t.Fatalf("Failed to get all cloud resources: %v", err)
		}

		if len(resp.Data) == 0 {
			t.Error("Expected data, got empty")
		}

		// Should have fetched all pages
		t.Logf("Fetched %d total resources", len(resp.Data))
	})
}
