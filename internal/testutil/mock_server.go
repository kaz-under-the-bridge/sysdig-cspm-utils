// Package testutil provides test fixtures and utilities for testing Sysdig CSPM API client.
package testutil

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// MockServerConfig holds configuration for the CSPM mock server
type MockServerConfig struct {
	// CompliancePageCount controls how many pages to return for compliance results
	CompliancePageCount int
	// CloudResourcesPageCount controls how many pages to return for cloud resources
	CloudResourcesPageCount int
	// DefaultPageSize is the default page size when not specified
	DefaultPageSize int
	// UnauthorizedResponse controls whether to return 401 for all requests
	UnauthorizedResponse bool
	// RateLimitResponse controls whether to return 429 for all requests
	RateLimitResponse bool
}

// DefaultMockServerConfig returns a default configuration
func DefaultMockServerConfig() *MockServerConfig {
	return &MockServerConfig{
		CompliancePageCount:     2,
		CloudResourcesPageCount: 2,
		DefaultPageSize:         10,
		UnauthorizedResponse:    false,
		RateLimitResponse:       false,
	}
}

// NewMockServer creates a new HTTP test server that mocks Sysdig CSPM API endpoints
func NewMockServer(config *MockServerConfig) *httptest.Server {
	if config == nil {
		config = DefaultMockServerConfig()
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 認証チェック
		if config.UnauthorizedResponse {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message": "unauthorized"}`))
			return
		}

		// レート制限チェック
		if config.RateLimitResponse {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "rate limit exceeded"}`))
			return
		}

		// Authorization ヘッダーチェック
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message": "missing authorization header"}`))
			return
		}

		// エンドポイントルーティング
		path := r.URL.Path

		switch {
		case strings.HasPrefix(path, "/api/cspm/v1/compliance/requirements"):
			handleComplianceRequirements(w, r, config)

		case strings.HasPrefix(path, "/api/cspm/v1/cloud/resources"):
			handleCloudResources(w, r, config)

		case strings.HasPrefix(path, "/api/cspm/v1/clusteranalysis/resources"):
			handleClusterAnalysisResources(w, r, config)

		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message": "endpoint not found"}`))
		}
	})

	return httptest.NewServer(handler)
}

// handleComplianceRequirements handles /api/cspm/v1/compliance/requirements
func handleComplianceRequirements(w http.ResponseWriter, r *http.Request, config *MockServerConfig) {
	w.Header().Set("Content-Type", "application/json")

	// クエリパラメータ解析
	query := r.URL.Query()
	pageNumberStr := query.Get("pageNumber")

	// ページ番号の解析
	pageNumber := 1
	if pageNumberStr != "" {
		if pn, err := strconv.Atoi(pageNumberStr); err == nil && pn > 0 {
			pageNumber = pn
		}
	}

	// ページ範囲チェック
	if pageNumber > config.CompliancePageCount {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"totalCount":"40"}`))
		return
	}

	// フィクスチャファイルを読み込み
	fixtureFile := fmt.Sprintf("fixtures/cloud_resources/compliance_requirements_page%d.json", pageNumber)
	data, err := LoadFixture(fixtureFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"message": "failed to load fixture: %v"}`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// handleCloudResources handles /api/cspm/v1/cloud/resources
func handleCloudResources(w http.ResponseWriter, r *http.Request, config *MockServerConfig) {
	w.Header().Set("Content-Type", "application/json")

	// クエリパラメータ解析
	query := r.URL.Query()
	controlID := query.Get("controlId")
	pageNumberStr := query.Get("pageNumber")

	// 必須パラメータチェック
	if controlID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "controlId parameter is required"}`))
		return
	}

	// ページ番号の解析
	pageNumber := 1
	if pageNumberStr != "" {
		if pn, err := strconv.Atoi(pageNumberStr); err == nil && pn > 0 {
			pageNumber = pn
		}
	}

	// ページ範囲チェック
	if pageNumber > config.CloudResourcesPageCount {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"totalCount":0}`))
		return
	}

	// コントロールIDに基づいてフィクスチャファイルを決定
	var fixtureFile string
	switch controlID {
	// CIS AWS (既存)
	case "16071":
		fixtureFile = fmt.Sprintf("fixtures/cloud_resources/control_16071_network_acl_page%d.json", pageNumber)
	case "16027":
		fixtureFile = fmt.Sprintf("fixtures/cloud_resources/control_16027_s3_mfa_delete_page%d.json", pageNumber)
	case "16026":
		fixtureFile = fmt.Sprintf("fixtures/cloud_resources/control_16026_s3_versioning_page%d.json", pageNumber)
	case "16018":
		fixtureFile = fmt.Sprintf("fixtures/cloud_resources/control_16018_iam_policy_page%d.json", pageNumber)
	// CIS GCP
	case "15041":
		fixtureFile = "fixtures/cloud_resources/cloud_resources_gcp_15041_vpc_flow_logs.json"
	case "15001":
		fixtureFile = "fixtures/cloud_resources/cloud_resources_gcp_15001_iam_mfa.json"
	// SOC2
	case "15024":
		fixtureFile = "fixtures/cloud_resources/cloud_resources_soc2_15024_gcp_project.json"
	default:
		// 未知のコントロールIDの場合は空レスポンス
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"totalCount":0}`))
		return
	}

	// フィクスチャファイルを読み込み
	data, err := LoadFixture(fixtureFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"message": "failed to load fixture: %v"}`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// handleClusterAnalysisResources handles /api/cspm/v1/clusteranalysis/resources
func handleClusterAnalysisResources(w http.ResponseWriter, r *http.Request, config *MockServerConfig) {
	w.Header().Set("Content-Type", "application/json")

	// クエリパラメータ解析
	query := r.URL.Query()
	controlID := query.Get("controlId")

	// 必須パラメータチェック
	if controlID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "controlId parameter is required"}`))
		return
	}

	// コントロールIDに基づいてフィクスチャファイルを決定
	var fixtureFile string
	switch controlID {
	case "5017":
		fixtureFile = "fixtures/cloud_resources/cluster_resources_soc2_5017_docker_host.json"
	case "6341":
		fixtureFile = "fixtures/cloud_resources/cluster_resources_soc2_6341_linux_host.json"
	default:
		// 未知のコントロールIDの場合は空レスポンス
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[],"totalCount":0}`))
		return
	}

	// フィクスチャファイルを読み込み
	data, err := LoadFixture(fixtureFile)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(fmt.Sprintf(`{"message": "failed to load fixture: %v"}`, err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}
