package testutil

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMockServer_ComplianceRequirements(t *testing.T) {
	server := NewMockServer(DefaultMockServerConfig())
	defer server.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "compliance requirements page 1",
			path:           "/api/cspm/v1/compliance/requirements?pageNumber=1&pageSize=2",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "compliance requirements page 2",
			path:           "/api/cspm/v1/compliance/requirements?pageNumber=2&pageSize=2",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "compliance requirements with CIS AWS filter (URL encoded)",
			path:           "/api/cspm/v1/compliance/requirements?filter=pass%3D%22false%22%20and%20policy.name%20in%20%28%22CIS%20Amazon%20Web%20Services%20Foundations%20Benchmark%20v3.0.0%22%29%20and%20zone.name%20in%20%28%22Entire%20Infrastructure%22%29",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "compliance requirements with CIS AWS filter (plus encoded)",
			path:           "/api/cspm/v1/compliance/requirements?filter=pass+%3D+%22false%22+and+policy.name+in+%28%22CIS+Amazon+Web+Services+Foundations+Benchmark+v3.0.0%22%29+and+zone.name+in+%28%22Entire+Infrastructure%22%29",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "compliance requirements page not found (beyond available pages)",
			path:           "/api/cspm/v1/compliance/requirements?pageNumber=999",
			expectedStatus: http.StatusOK,
			checkResponse:  false, // Empty response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Authorization", "Bearer test-token")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Check if response has expected structure
				if _, ok := response["data"]; !ok {
					t.Error("Response should have 'data' field")
				}
				if _, ok := response["totalCount"]; !ok {
					t.Error("Response should have 'totalCount' field")
				}

				// Check for policy-specific responses
				if data, ok := response["data"].([]interface{}); ok && len(data) > 0 {
					firstItem := data[0].(map[string]interface{})
					policyName := firstItem["policyName"].(string)

					if tt.name == "compliance requirements with CIS AWS filter (URL encoded)" ||
						tt.name == "compliance requirements with CIS AWS filter (plus encoded)" {
						if policyName != "CIS Amazon Web Services Foundations Benchmark v3.0.0" {
							t.Errorf("Expected CIS AWS policy name, got %s", policyName)
						}
					}
				}
			}
		})
	}
}

func TestMockServer_CloudResources(t *testing.T) {
	server := NewMockServer(DefaultMockServerConfig())
	defer server.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "cloud resources control 16071 page 1",
			path:           "/api/cspm/v1/cloud/resources?controlId=16071&pageNumber=1&pageSize=2",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "cloud resources control 16071 page 2",
			path:           "/api/cspm/v1/cloud/resources?controlId=16071&pageNumber=2&pageSize=2",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "cloud resources control 16027 page 1",
			path:           "/api/cspm/v1/cloud/resources?controlId=16027&pageNumber=1",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "cloud resources control 16026 page 1",
			path:           "/api/cspm/v1/cloud/resources?controlId=16026&pageNumber=1",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "cloud resources control 16018 page 1",
			path:           "/api/cspm/v1/cloud/resources?controlId=16018&pageNumber=1",
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "cloud resources missing controlId",
			path:           "/api/cspm/v1/cloud/resources?pageNumber=1",
			expectedStatus: http.StatusBadRequest,
			checkResponse:  false,
		},
		{
			name:           "cloud resources unknown controlId",
			path:           "/api/cspm/v1/cloud/resources?controlId=99999&pageNumber=1",
			expectedStatus: http.StatusOK,
			checkResponse:  false, // Empty response
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL+tt.path, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Authorization", "Bearer test-token")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.checkResponse {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Fatalf("Failed to read response body: %v", err)
				}

				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to parse JSON response: %v", err)
				}

				// Check if response has expected structure
				if _, ok := response["data"]; !ok {
					t.Error("Response should have 'data' field")
				}
				if _, ok := response["totalCount"]; !ok {
					t.Error("Response should have 'totalCount' field")
				}
			}
		})
	}
}

func TestMockServer_Authentication(t *testing.T) {
	tests := []struct {
		name           string
		serverFunc     func() *httptest.Server
		authHeader     string
		expectedStatus int
	}{
		{
			name: "missing authorization header",
			serverFunc: func() *httptest.Server {
				return NewMockServer(DefaultMockServerConfig())
			},
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid authorization header",
			serverFunc: func() *httptest.Server {
				return NewMockServer(DefaultMockServerConfig())
			},
			authHeader:     "Bearer test-token",
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized server",
			serverFunc: func() *httptest.Server {
				config := DefaultMockServerConfig()
				config.UnauthorizedResponse = true
				return NewMockServer(config)
			},
			authHeader:     "Bearer test-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "rate limited server",
			serverFunc: func() *httptest.Server {
				config := DefaultMockServerConfig()
				config.RateLimitResponse = true
				return NewMockServer(config)
			},
			authHeader:     "Bearer test-token",
			expectedStatus: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.serverFunc()
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL+"/api/cspm/v1/compliance/requirements", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestMockServer_UnknownEndpoint(t *testing.T) {
	server := NewMockServer(DefaultMockServerConfig())
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL+"/unknown/endpoint", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer test-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
