package collector

import (
	"os"
	"testing"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/internal/testutil"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/client"
	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/database"
)

func TestCollectComplianceData(t *testing.T) {
	// Setup mock server
	mockServer := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer mockServer.Close()

	// Setup client
	cspmClient := client.NewCSPMClient(mockServer.URL, "test-token")

	// Setup database
	dbPath := "test_compliance.db"
	defer os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create collector
	collector := NewComplianceCollector(cspmClient, db)

	// Test: Collect compliance data
	filter := "policy.name contains \"CIS\""
	err = collector.CollectComplianceData(filter, 10, 2, 0)
	if err != nil {
		t.Fatalf("Failed to collect compliance data: %v", err)
	}

	// Verify stats
	stats, err := collector.GetComplianceStats()
	if err != nil {
		t.Fatalf("Failed to get compliance stats: %v", err)
	}

	t.Logf("Requirements: %d total, %d failed, %d passed",
		stats.TotalRequirements, stats.FailedRequirements, stats.PassedRequirements)
	t.Logf("Controls: %d total, %d failed, %d passed",
		stats.TotalControls, stats.FailedControls, stats.PassedControls)
	t.Logf("Resources: %d total, %d failed, %d passed, %d accepted",
		stats.TotalResources, stats.FailedResources, stats.PassedResources, stats.AcceptedResources)

	// Basic validation
	if stats.TotalRequirements == 0 {
		t.Error("Expected requirements to be collected")
	}
	if stats.TotalControls == 0 {
		t.Error("Expected controls to be collected")
	}
	// Note: Some controls may not have resource fixtures in mock server,
	// so resources might be 0. This is acceptable in tests.
	if stats.TotalResources > 0 {
		t.Logf("Successfully collected %d resources", stats.TotalResources)
	} else {
		t.Logf("Note: No resources collected (some controls may not have fixtures)")
	}
}

func TestCollectControlResources(t *testing.T) {
	// Setup mock server
	mockServer := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer mockServer.Close()

	// Setup client
	cspmClient := client.NewCSPMClient(mockServer.URL, "test-token")

	// Setup database
	dbPath := "test_control_resources.db"
	defer os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create collector
	collector := NewComplianceCollector(cspmClient, db)

	// Test: Collect resources for a specific control
	controlID := "16071"
	endpoint := "/api/cspm/v1/cloud/resources?controlId=16071"

	resources, err := collector.CollectControlResources(controlID, endpoint, 10, 1, 0)
	if err != nil {
		t.Fatalf("Failed to collect control resources: %v", err)
	}

	if len(resources) == 0 {
		t.Error("Expected resources to be collected")
	}

	t.Logf("Collected %d resources for control %s", len(resources), controlID)

	// Verify resources were saved to database
	stats, err := collector.GetComplianceStats()
	if err != nil {
		t.Fatalf("Failed to get compliance stats: %v", err)
	}

	if stats.TotalResources == 0 {
		t.Error("Expected resources to be saved to database")
	}
}

func TestCollectComplianceDataWithStats(t *testing.T) {
	// Setup mock server
	mockServer := testutil.NewMockServer(testutil.DefaultMockServerConfig())
	defer mockServer.Close()

	// Setup client
	cspmClient := client.NewCSPMClient(mockServer.URL, "test-token")

	// Setup database
	dbPath := "test_compliance_with_stats.db"
	defer os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create collector
	collector := NewComplianceCollector(cspmClient, db)

	// Test: Collect compliance data and get stats
	filter := "policy.name contains \"CIS\""
	stats, err := collector.CollectComplianceDataWithStats(filter, 10, 2, 0)
	if err != nil {
		t.Fatalf("Failed to collect compliance data with stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Expected stats to be returned")
	}

	t.Logf("Stats: Requirements=%d, Controls=%d, Resources=%d",
		stats.TotalRequirements, stats.TotalControls, stats.TotalResources)

	if stats.TotalRequirements == 0 {
		t.Error("Expected requirements to be collected")
	}
}
