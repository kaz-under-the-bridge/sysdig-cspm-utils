package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/kaz-under-the-bridge/sysdig-cspm-utils/pkg/models"
)

// SaveComplianceRequirements saves compliance requirements to database
func (d *Database) SaveComplianceRequirements(requirements []models.ComplianceRequirement) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO compliance_requirements (
			requirement_id, name, policy_id, policy_name, policy_type, platform,
			severity, pass, zone_id, zone_name, failed_controls,
			high_severity_count, medium_severity_count, low_severity_count,
			accepted_count, passing_count, description, resource_api_endpoint
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, req := range requirements {
		_, err = stmt.Exec(
			req.RequirementID, req.Name, req.PolicyID, req.PolicyName, req.PolicyType, req.Platform,
			req.Severity, req.Pass, req.ZoneID, req.ZoneName, req.FailedControls,
			req.HighSeverityCount, req.MediumSeverityCount, req.LowSeverityCount,
			req.AcceptedCount, req.PassingCount, req.Description, req.ResourceAPIEndpoint,
		)
		if err != nil {
			return fmt.Errorf("failed to insert compliance requirement: %w", err)
		}
	}

	return tx.Commit()
}

// GetComplianceViolations retrieves compliance violations from database
func (d *Database) GetComplianceViolations(policyType string, platform string) ([]models.ComplianceRequirement, error) {
	query := `
		SELECT requirement_id, name, policy_id, policy_name, policy_type, platform,
		       severity, pass, zone_id, zone_name, failed_controls,
		       high_severity_count, medium_severity_count, low_severity_count,
		       accepted_count, passing_count, description, resource_api_endpoint
		FROM compliance_requirements
		WHERE pass = false`

	args := []interface{}{}

	if policyType != "" {
		query += " AND policy_type = ?"
		args = append(args, policyType)
	}

	if platform != "" {
		query += " AND platform = ?"
		args = append(args, platform)
	}

	query += " ORDER BY severity DESC, failed_controls DESC"

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query compliance violations: %w", err)
	}
	defer rows.Close()

	var violations []models.ComplianceRequirement
	for rows.Next() {
		var req models.ComplianceRequirement
		err = rows.Scan(
			&req.RequirementID, &req.Name, &req.PolicyID, &req.PolicyName, &req.PolicyType, &req.Platform,
			&req.Severity, &req.Pass, &req.ZoneID, &req.ZoneName, &req.FailedControls,
			&req.HighSeverityCount, &req.MediumSeverityCount, &req.LowSeverityCount,
			&req.AcceptedCount, &req.PassingCount, &req.Description, &req.ResourceAPIEndpoint,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan compliance requirement: %w", err)
		}
		violations = append(violations, req)
	}

	return violations, nil
}

// SaveComplianceRequirementsWithControls saves compliance requirements and their controls to the database
func (d *Database) SaveComplianceRequirementsWithControls(requirements []models.ComplianceRequirementWithControls) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statements
	reqStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO compliance_requirements (
			requirement_id, name, policy_id, policy_name, policy_type, platform,
			severity, pass, zone_id, zone_name, failed_controls, description
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare requirement statement: %w", err)
	}
	defer reqStmt.Close()

	ctrlStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO controls (
			control_id, name, description, requirement_id, severity, pass,
			objects_count, passing_count, accepted_count, resource_kind,
			resource_api_endpoint, target, platform
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare control statement: %w", err)
	}
	defer ctrlStmt.Close()

	// Insert requirements and controls
	for _, req := range requirements {
		// Extract zone info
		zoneID := req.Zone.ID
		zoneName := req.Zone.Name

		// Extract policy type from policy name (heuristic)
		policyType := extractPolicyType(req.PolicyName)

		// Extract platform from policy name (heuristic)
		platform := extractPlatform(req.PolicyName)

		// Insert requirement
		_, err := reqStmt.Exec(
			req.RequirementID,
			req.Name,
			req.PolicyID,
			req.PolicyName,
			policyType,
			platform,
			req.Severity,
			req.Pass,
			zoneID,
			zoneName,
			req.FailedControls,
			req.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to insert requirement %s: %w", req.RequirementID, err)
		}

		// Insert controls
		for _, ctrl := range req.Controls {
			_, err := ctrlStmt.Exec(
				ctrl.ID,
				ctrl.Name,
				ctrl.Description,
				req.RequirementID,
				ctrl.Severity,
				ctrl.Pass,
				ctrl.ObjectsCount,
				ctrl.PassingCount,
				ctrl.AcceptedCount,
				ctrl.ResourceKind,
				ctrl.ResourceAPIEndpoint,
				ctrl.Target,
				ctrl.Platform,
			)
			if err != nil {
				return fmt.Errorf("failed to insert control %s: %w", ctrl.ID, err)
			}
		}
	}

	return tx.Commit()
}

// SaveCloudResources saves cloud resources to the database
func (d *Database) SaveCloudResources(resources []models.CloudResource) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO cloud_resources (
			hash, name, type, platform, account, location, organization,
			os_name, os_image, cluster_name, distribution_name, distribution_version,
			zones_json, label_values_json, last_seen_date, global_id,
			platform_account_id, cloud_resource_id, cloud_region, agent_tags_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, res := range resources {
		// Serialize JSON fields
		zonesJSON, err := json.Marshal(res.Zones)
		if err != nil {
			return fmt.Errorf("failed to marshal zones for %s: %w", res.Hash, err)
		}

		labelValuesJSON, err := json.Marshal(res.LabelValues)
		if err != nil {
			return fmt.Errorf("failed to marshal label values for %s: %w", res.Hash, err)
		}

		agentTagsJSON, err := json.Marshal(res.AgentTags)
		if err != nil {
			return fmt.Errorf("failed to marshal agent tags for %s: %w", res.Hash, err)
		}

		_, err = stmt.Exec(
			res.Hash,
			res.Name,
			res.Type,
			nullString(res.Platform),
			nullString(res.Account),
			nullString(res.Location),
			nullString(res.Organization),
			nullString(res.OSName),
			nullString(res.OSImage),
			nullString(res.ClusterName),
			nullString(res.DistributionName),
			nullString(res.DistributionVersion),
			string(zonesJSON),
			string(labelValuesJSON),
			res.LastSeenDate,
			res.GlobalID,
			nullString(res.PlatformAccountID),
			nullString(res.CloudResourceID),
			nullString(res.CloudRegion),
			string(agentTagsJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to insert resource %s: %w", res.Hash, err)
		}
	}

	return tx.Commit()
}

// SaveControlResourceRelations saves control-resource relationships to the database
func (d *Database) SaveControlResourceRelations(controlID string, resources []models.CloudResource) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO control_resource_relations (
			control_id, resource_hash, passed, acceptance_status, acceptance_justification
		) VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, res := range resources {
		acceptanceStatus := res.GetAcceptanceStatus()
		var justification sql.NullString
		if res.Acceptance != nil {
			justification = sql.NullString{String: res.Acceptance.Justification, Valid: true}
		}

		_, err := stmt.Exec(
			controlID,
			res.Hash,
			res.Passed,
			acceptanceStatus,
			justification,
		)
		if err != nil {
			return fmt.Errorf("failed to insert relation for control %s, resource %s: %w", controlID, res.Hash, err)
		}
	}

	return tx.Commit()
}

// GetComplianceStats returns statistics about compliance requirements
func (d *Database) GetComplianceStats() (*ComplianceStats, error) {
	var stats ComplianceStats

	// Get requirement stats
	err := d.db.QueryRow(`
		SELECT
			COUNT(*) as total_requirements,
			COUNT(CASE WHEN pass = 0 THEN 1 END) as failed_requirements,
			COUNT(CASE WHEN pass = 1 THEN 1 END) as passed_requirements
		FROM compliance_requirements
	`).Scan(&stats.TotalRequirements, &stats.FailedRequirements, &stats.PassedRequirements)
	if err != nil {
		return nil, fmt.Errorf("failed to get requirement stats: %w", err)
	}

	// Get control stats
	err = d.db.QueryRow(`
		SELECT
			COUNT(*) as total_controls,
			COUNT(CASE WHEN pass = 0 THEN 1 END) as failed_controls,
			COUNT(CASE WHEN pass = 1 THEN 1 END) as passed_controls
		FROM controls
	`).Scan(&stats.TotalControls, &stats.FailedControls, &stats.PassedControls)
	if err != nil {
		return nil, fmt.Errorf("failed to get control stats: %w", err)
	}

	// Get resource stats
	err = d.db.QueryRow(`
		SELECT
			COUNT(DISTINCT resource_hash) as total_resources,
			COUNT(CASE WHEN acceptance_status = 'failed' THEN 1 END) as failed_resources,
			COUNT(CASE WHEN acceptance_status = 'passed' THEN 1 END) as passed_resources,
			COUNT(CASE WHEN acceptance_status = 'accepted' THEN 1 END) as accepted_resources
		FROM control_resource_relations
	`).Scan(&stats.TotalResources, &stats.FailedResources, &stats.PassedResources, &stats.AcceptedResources)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource stats: %w", err)
	}

	return &stats, nil
}

// ComplianceStats holds statistics about compliance data
type ComplianceStats struct {
	TotalRequirements  int
	FailedRequirements int
	PassedRequirements int
	TotalControls      int
	FailedControls     int
	PassedControls     int
	TotalResources     int
	FailedResources    int
	PassedResources    int
	AcceptedResources  int
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// extractPolicyType extracts policy type from policy name
func extractPolicyType(policyName string) string {
	// Heuristic: Check for common policy type keywords
	switch {
	case contains(policyName, "CIS"):
		return "CIS"
	case contains(policyName, "SOC"):
		return "SOC2"
	case contains(policyName, "SOC 2"):
		return "SOC2"
	case contains(policyName, "PCI"):
		return "PCI-DSS"
	case contains(policyName, "HIPAA"):
		return "HIPAA"
	case contains(policyName, "NIST"):
		return "NIST"
	default:
		return "Unknown"
	}
}

// extractPlatform extracts platform from policy name
func extractPlatform(policyName string) string {
	// Heuristic: Check for platform keywords
	switch {
	case contains(policyName, "AWS"):
		return "AWS"
	case contains(policyName, "Amazon"):
		return "AWS"
	case contains(policyName, "GCP"):
		return "GCP"
	case contains(policyName, "Google Cloud"):
		return "GCP"
	case contains(policyName, "Azure"):
		return "Azure"
	case contains(policyName, "Kubernetes"):
		return "Kubernetes"
	case contains(policyName, "K8s"):
		return "Kubernetes"
	default:
		return "Multi-Cloud"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(sLower) >= len(substrLower) && findSubstring(sLower, substrLower)
}

// toLower converts a string to lowercase
func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + 32
		} else {
			result[i] = r
		}
	}
	return string(result)
}

// findSubstring checks if haystack contains needle
func findSubstring(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
