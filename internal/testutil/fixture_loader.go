// Package testutil provides test fixtures and utilities for testing Sysdig CSPM API client.
package testutil

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// LoadFixture loads a fixture file from the fixtures directory
func LoadFixture(filename string) ([]byte, error) {
	// 現在のファイルのディレクトリを取得
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	// internal/testutil/fixtures/ へのパスを構築
	baseDir := filepath.Dir(currentFile)
	fixturePath := filepath.Join(baseDir, filename)

	// ファイルを読み込み
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file %s: %w", fixturePath, err)
	}

	return data, nil
}

// ComplianceRequirementsPage1 returns the first page of compliance requirements
func ComplianceRequirementsPage1() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/compliance_requirements_page1.json")
}

// ComplianceRequirementsPage2 returns the second page of compliance requirements
func ComplianceRequirementsPage2() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/compliance_requirements_page2.json")
}

// CloudResourcesControl16071Page1 returns the first page of Network ACL resources
func CloudResourcesControl16071Page1() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16071_network_acl_page1.json")
}

// CloudResourcesControl16071Page2 returns the second page of Network ACL resources
func CloudResourcesControl16071Page2() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16071_network_acl_page2.json")
}

// CloudResourcesControl16027Page1 returns the first page of S3 MFA Delete resources
func CloudResourcesControl16027Page1() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16027_s3_mfa_delete_page1.json")
}

// CloudResourcesControl16027Page2 returns the second page of S3 MFA Delete resources
func CloudResourcesControl16027Page2() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16027_s3_mfa_delete_page2.json")
}

// CloudResourcesControl16026Page1 returns the first page of S3 Versioning resources
func CloudResourcesControl16026Page1() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16026_s3_versioning_page1.json")
}

// CloudResourcesControl16026Page2 returns the second page of S3 Versioning resources
func CloudResourcesControl16026Page2() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16026_s3_versioning_page2.json")
}

// CloudResourcesControl16018Page1 returns the first page of IAM Policy resources
func CloudResourcesControl16018Page1() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16018_iam_policy_page1.json")
}

// CloudResourcesControl16018Page2 returns the second page of IAM Policy resources
func CloudResourcesControl16018Page2() ([]byte, error) {
	return LoadFixture("fixtures/cloud_resources/control_16018_iam_policy_page2.json")
}
