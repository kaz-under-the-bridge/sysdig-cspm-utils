package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/models"
)

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name      string
		dbPath    string
		wantError bool
	}{
		{
			name:      "有効なデータベース作成",
			dbPath:    filepath.Join(t.TempDir(), "test.db"),
			wantError: false,
		},
		{
			name: "ネストされたディレクトリ",
			dbPath: func() string {
				dir := filepath.Join(t.TempDir(), "subdir")
				os.MkdirAll(dir, 0755)
				return filepath.Join(dir, "test.db")
			}(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewDatabase(tt.dbPath)

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

			if db == nil {
				t.Error("Database is nil")
				return
			}

			defer db.Close()

			// データベースファイルが作成されたか確認
			if _, err := os.Stat(tt.dbPath); os.IsNotExist(err) {
				t.Errorf("Database file was not created: %s", tt.dbPath)
			}
		})
	}
}

func TestSaveComplianceRequirements(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	requirements := []models.ComplianceRequirement{
		{
			RequirementID:       "req-1",
			Name:                "Test Requirement 1",
			PolicyID:            "policy-1",
			PolicyName:          "Test Policy",
			PolicyType:          "CIS",
			Platform:            "AWS",
			Severity:            "High",
			Pass:                false,
			ZoneID:              "zone-1",
			ZoneName:            "Entire Infrastructure",
			FailedControls:      5,
			HighSeverityCount:   2,
			MediumSeverityCount: 2,
			LowSeverityCount:    1,
			AcceptedCount:       0,
			PassingCount:        10,
			Description:         "Test description",
			ResourceAPIEndpoint: "/api/endpoint",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
		{
			RequirementID:       "req-2",
			Name:                "Test Requirement 2",
			PolicyID:            "policy-1",
			PolicyName:          "Test Policy",
			PolicyType:          "CIS",
			Platform:            "GCP",
			Severity:            "Medium",
			Pass:                false,
			ZoneID:              "zone-1",
			ZoneName:            "Entire Infrastructure",
			FailedControls:      3,
			HighSeverityCount:   0,
			MediumSeverityCount: 2,
			LowSeverityCount:    1,
			AcceptedCount:       0,
			PassingCount:        15,
			Description:         "Test description 2",
			ResourceAPIEndpoint: "/api/endpoint2",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		},
	}

	err = db.SaveComplianceRequirements(requirements)
	if err != nil {
		t.Errorf("Failed to save compliance requirements: %v", err)
	}

	// 保存されたデータを確認
	rows, err := db.db.Query("SELECT requirement_id, name, platform FROM compliance_requirements")
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var reqID, name, platform string
		if err := rows.Scan(&reqID, &name, &platform); err != nil {
			t.Errorf("Failed to scan row: %v", err)
		}
		count++
	}

	if count != len(requirements) {
		t.Errorf("Expected %d requirements, got %d", len(requirements), count)
	}
}

func TestDatabaseClose(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}
}

func TestSaveComplianceRequirements_DuplicateHandling(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	requirement := models.ComplianceRequirement{
		RequirementID: "req-dup",
		Name:          "Original Name",
		PolicyID:      "policy-1",
		PolicyName:    "Test Policy",
		Platform:      "AWS",
		Severity:      "High",
		Pass:          false,
		ZoneID:        "zone-1",
		ZoneName:      "Entire Infrastructure",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 最初の保存
	err = db.SaveComplianceRequirements([]models.ComplianceRequirement{requirement})
	if err != nil {
		t.Errorf("Failed to save first requirement: %v", err)
	}

	// 同じIDで更新
	requirement.Name = "Updated Name"
	err = db.SaveComplianceRequirements([]models.ComplianceRequirement{requirement})
	if err != nil {
		t.Errorf("Failed to save updated requirement: %v", err)
	}

	// 更新されたことを確認
	var name string
	err = db.db.QueryRow("SELECT name FROM compliance_requirements WHERE requirement_id = ?", "req-dup").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if name != "Updated Name" {
		t.Errorf("Expected 'Updated Name', got '%s'", name)
	}
}

func TestGetComplianceViolations(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// テストデータを準備
	requirements := []models.ComplianceRequirement{
		{
			RequirementID:  "req-aws-1",
			Name:           "AWS Test Requirement",
			PolicyID:       "policy-1",
			PolicyName:     "Test Policy",
			PolicyType:     "CIS",
			Platform:       "AWS",
			Severity:       "High",
			Pass:           false,
			ZoneID:         "zone-1",
			ZoneName:       "Entire Infrastructure",
			FailedControls: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			RequirementID:  "req-gcp-1",
			Name:           "GCP Test Requirement",
			PolicyID:       "policy-2",
			PolicyName:     "Test Policy 2",
			PolicyType:     "SOC2",
			Platform:       "GCP",
			Severity:       "Medium",
			Pass:           false,
			ZoneID:         "zone-1",
			ZoneName:       "Entire Infrastructure",
			FailedControls: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	err = db.SaveComplianceRequirements(requirements)
	if err != nil {
		t.Fatalf("Failed to save requirements: %v", err)
	}

	tests := []struct {
		name         string
		policyType   string
		platform     string
		wantCount    int
		wantPlatform string
	}{
		{
			name:         "全ての違反",
			policyType:   "",
			platform:     "",
			wantCount:    2,
			wantPlatform: "",
		},
		{
			name:         "AWSのみ",
			policyType:   "",
			platform:     "AWS",
			wantCount:    1,
			wantPlatform: "AWS",
		},
		{
			name:         "CISポリシーのみ",
			policyType:   "CIS",
			platform:     "",
			wantCount:    1,
			wantPlatform: "AWS",
		},
		{
			name:         "CIS + AWS",
			policyType:   "CIS",
			platform:     "AWS",
			wantCount:    1,
			wantPlatform: "AWS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations, err := db.GetComplianceViolations(tt.policyType, tt.platform)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(violations) != tt.wantCount {
				t.Errorf("Expected %d violations, got %d", tt.wantCount, len(violations))
				return
			}

			if tt.wantPlatform != "" && len(violations) > 0 {
				if violations[0].Platform != tt.wantPlatform {
					t.Errorf("Expected platform %s, got %s", tt.wantPlatform, violations[0].Platform)
				}
			}
		})
	}
}

func TestDB_Getter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	sqlDB := db.DB()
	if sqlDB == nil {
		t.Error("DB() returned nil")
	}

	// データベースが使用可能か確認
	err = sqlDB.Ping()
	if err != nil {
		t.Errorf("Database ping failed: %v", err)
	}
}

func TestSaveComplianceRequirementsWithControls(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	requirements := []models.ComplianceRequirementWithControls{
		{
			RequirementID:  "req-1",
			Name:           "Test Requirement",
			PolicyID:       "policy-1",
			PolicyName:     "CIS Amazon Web Services Foundations Benchmark v3.0.0",
			Severity:       "High",
			Pass:           false,
			FailedControls: 2,
			Description:    "Test description",
			Zone: models.Zone{
				ID:   "zone-1",
				Name: "Test Zone",
			},
			Controls: []models.Control{
				{
					ID:                  "ctrl-1",
					Name:                "Test Control 1",
					Description:         "Control description",
					Severity:            "High",
					Pass:                false,
					ObjectsCount:        5,
					PassingCount:        0,
					AcceptedCount:       0,
					ResourceKind:        "aws_s3_bucket",
					ResourceAPIEndpoint: "/api/resources",
					Target:              "AWS",
					Platform:            "AWS",
				},
			},
		},
	}

	err = db.SaveComplianceRequirementsWithControls(requirements)
	if err != nil {
		t.Errorf("Failed to save compliance requirements with controls: %v", err)
	}

	// 保存されたデータを確認
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM compliance_requirements").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query requirements: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 requirement, got %d", count)
	}

	err = db.db.QueryRow("SELECT COUNT(*) FROM controls").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query controls: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 control, got %d", count)
	}
}

func TestSaveCloudResources(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	resources := []models.CloudResource{
		{
			Hash:         "hash-1",
			Name:         "test-resource-1",
			Type:         "aws_s3_bucket",
			Platform:     "AWS",
			Account:      "123456789012",
			Location:     "us-east-1",
			Zones:        []models.Zone{{ID: "zone-1", Name: "Zone 1"}},
			LabelValues:  []string{"env:prod"},
			LastSeenDate: "1234567890",
			GlobalID:     "global-id-1",
		},
		{
			Hash:         "hash-2",
			Name:         "test-resource-2",
			Type:         "gcp_storage_bucket",
			Platform:     "GCP",
			Account:      "my-project",
			Location:     "us-central1",
			Zones:        []models.Zone{{ID: "zone-1", Name: "Zone 1"}},
			LabelValues:  []string{"env:dev"},
			LastSeenDate: "1234567891",
			GlobalID:     "global-id-2",
		},
	}

	err = db.SaveCloudResources(resources)
	if err != nil {
		t.Errorf("Failed to save cloud resources: %v", err)
	}

	// 保存されたデータを確認
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM cloud_resources").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query resources: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 resources, got %d", count)
	}
}

func TestSaveControlResourceRelations(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// まずリソースを保存
	resources := []models.CloudResource{
		{
			Hash:         "hash-1",
			Name:         "test-resource-1",
			Type:         "aws_s3_bucket",
			Passed:       false,
			LastSeenDate: "1234567890",
			GlobalID:     "global-id-1",
		},
		{
			Hash:         "hash-2",
			Name:         "test-resource-2",
			Type:         "aws_s3_bucket",
			Passed:       true,
			LastSeenDate: "1234567891",
			GlobalID:     "global-id-2",
		},
	}

	err = db.SaveCloudResources(resources)
	if err != nil {
		t.Fatalf("Failed to save resources: %v", err)
	}

	// コントロール-リソース関連を保存
	err = db.SaveControlResourceRelations("ctrl-1", resources)
	if err != nil {
		t.Errorf("Failed to save control-resource relations: %v", err)
	}

	// 保存されたデータを確認
	var count int
	err = db.db.QueryRow("SELECT COUNT(*) FROM control_resource_relations").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query relations: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 relations, got %d", count)
	}
}

func TestGetComplianceStats(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// テストデータを準備
	requirements := []models.ComplianceRequirementWithControls{
		{
			RequirementID:  "req-1",
			Name:           "Failed Requirement",
			PolicyID:       "policy-1",
			PolicyName:     "Test Policy",
			Pass:           false,
			FailedControls: 1,
			Zone: models.Zone{
				ID:   "zone-1",
				Name: "Test Zone",
			},
			Controls: []models.Control{
				{
					ID:           "ctrl-1",
					Name:         "Failed Control",
					Pass:         false,
					ObjectsCount: 5,
					PassingCount: 3,
				},
			},
		},
		{
			RequirementID: "req-2",
			Name:          "Passed Requirement",
			PolicyID:      "policy-1",
			PolicyName:    "Test Policy",
			Pass:          true,
			Zone: models.Zone{
				ID:   "zone-1",
				Name: "Test Zone",
			},
			Controls: []models.Control{
				{
					ID:           "ctrl-2",
					Name:         "Passed Control",
					Pass:         true,
					ObjectsCount: 0,
					PassingCount: 5,
				},
			},
		},
	}

	err = db.SaveComplianceRequirementsWithControls(requirements)
	if err != nil {
		t.Fatalf("Failed to save requirements: %v", err)
	}

	// リソースと関連を保存
	resources := []models.CloudResource{
		{Hash: "hash-1", Name: "res-1", Type: "type-1", Passed: false, LastSeenDate: "123", GlobalID: "g1"},
		{Hash: "hash-2", Name: "res-2", Type: "type-1", Passed: true, LastSeenDate: "123", GlobalID: "g2"},
		{Hash: "hash-3", Name: "res-3", Type: "type-1", Passed: false, LastSeenDate: "123", GlobalID: "g3"},
	}
	err = db.SaveCloudResources(resources)
	if err != nil {
		t.Fatalf("Failed to save resources: %v", err)
	}

	err = db.SaveControlResourceRelations("ctrl-1", resources)
	if err != nil {
		t.Fatalf("Failed to save relations: %v", err)
	}

	// 統計を取得
	stats, err := db.GetComplianceStats()
	if err != nil {
		t.Fatalf("Failed to get compliance stats: %v", err)
	}

	// 統計を検証
	if stats.TotalRequirements != 2 {
		t.Errorf("Expected 2 total requirements, got %d", stats.TotalRequirements)
	}
	if stats.FailedRequirements != 1 {
		t.Errorf("Expected 1 failed requirement, got %d", stats.FailedRequirements)
	}
	if stats.PassedRequirements != 1 {
		t.Errorf("Expected 1 passed requirement, got %d", stats.PassedRequirements)
	}
	if stats.TotalControls != 2 {
		t.Errorf("Expected 2 total controls, got %d", stats.TotalControls)
	}
	if stats.TotalResources != 3 {
		t.Errorf("Expected 3 total resources, got %d", stats.TotalResources)
	}
}

func TestExtractPolicyType(t *testing.T) {
	tests := []struct {
		name       string
		policyName string
		want       string
	}{
		{"CIS policy", "CIS Amazon Web Services", "CIS"},
		{"SOC2 policy", "SOC 2", "SOC2"},
		{"PCI policy", "PCI-DSS v3.2", "PCI-DSS"},
		{"Unknown policy", "Custom Policy", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPolicyType(tt.policyName)
			if got != tt.want {
				t.Errorf("extractPolicyType(%q) = %q, want %q", tt.policyName, got, tt.want)
			}
		})
	}
}

func TestExtractPlatform(t *testing.T) {
	tests := []struct {
		name       string
		policyName string
		want       string
	}{
		{"AWS policy", "CIS Amazon Web Services", "AWS"},
		{"GCP policy", "CIS Google Cloud Platform", "GCP"},
		{"Azure policy", "Azure Security Benchmark", "Azure"},
		{"Multi-cloud", "Custom Policy", "Multi-Cloud"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPlatform(tt.policyName)
			if got != tt.want {
				t.Errorf("extractPlatform(%q) = %q, want %q", tt.policyName, got, tt.want)
			}
		})
	}
}
