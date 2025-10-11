package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// ComplianceRequirement represents a compliance requirement with violations
type ComplianceRequirement struct {
	RequirementID       string    `json:"requirementId"`
	Name                string    `json:"name"`
	PolicyID            string    `json:"policyId"`
	PolicyName          string    `json:"policyName"`
	PolicyType          string    `json:"policyType"`
	Platform            string    `json:"platform"`
	Severity            string    `json:"severity"`
	Pass                bool      `json:"pass"`
	ZoneID              string    `json:"zoneId"`
	ZoneName            string    `json:"zoneName"`
	FailedControls      int       `json:"failedControls"`
	HighSeverityCount   int       `json:"highSeverityCount"`
	MediumSeverityCount int       `json:"mediumSeverityCount"`
	LowSeverityCount    int       `json:"lowSeverityCount"`
	AcceptedCount       int       `json:"acceptedCount"`
	PassingCount        int       `json:"passingCount"`
	Description         string    `json:"description"`
	ResourceAPIEndpoint string    `json:"resourceApiEndpoint"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// InventoryResource represents a resource from Sysdig Inventory API
type InventoryResource struct {
	Hash                  string                 `json:"hash"`
	Name                  string                 `json:"name"`
	Type                  string                 `json:"type"`
	Platform              string                 `json:"platform"`
	Category              string                 `json:"category"`
	LastSeen              string                 `json:"lastSeen"`
	Metadata              map[string]interface{} `json:"metadata"`
	PosturePolicySummary  PolicySummary          `json:"posturePolicySummary"`
	Labels                []string               `json:"labels"`
	PostureControlSummary []ControlSummary       `json:"postureControlSummary"`
	Zones                 []Zone                 `json:"zones"`
	ResourceOrigin        string                 `json:"resourceOrigin"`
	ConfigAPIEndpoint     string                 `json:"configApiEndpoint"`
	CreatedAt             time.Time              `json:"createdAt"`
	UpdatedAt             time.Time              `json:"updatedAt"`
}

// PolicySummary represents posture policy summary
type PolicySummary struct {
	PassPercentage float64  `json:"passPercentage"`
	Policies       []Policy `json:"policies"`
}

// Policy represents a security policy
type Policy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Pass bool   `json:"pass"`
}

// ControlSummary represents posture control summary
type ControlSummary struct {
	Name             string `json:"name"`
	PolicyID         string `json:"policyId"`
	FailedControls   int    `json:"failedControls"`
	AcceptedControls int    `json:"acceptedControls"`
}

// Zone represents a zone that a resource belongs to
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Control represents a control within a compliance requirement
type Control struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	Target              string `json:"target"`
	Type                int    `json:"type"`
	Pass                bool   `json:"pass"`
	Severity            string `json:"severity"`
	ObjectsCount        int    `json:"objectsCount"`  // Failed リソース数
	PassingCount        int    `json:"passingCount"`  // Passed リソース数
	AcceptedCount       int    `json:"acceptedCount"` // Accepted リソース数
	RemediationID       string `json:"remediationId"`
	LastUpdate          string `json:"lastUpdate"`
	ResourceKind        string `json:"resourceKind"`
	IsManual            bool   `json:"isManual"`
	ResourceAPIEndpoint string `json:"resourceApiEndpoint"` // Cloud Resources API URL
	Platform            string `json:"platform"`
	Authors             string `json:"authors"`
}

// ComplianceResponse represents the API response for compliance requirements
type ComplianceResponse struct {
	Data       []ComplianceRequirement `json:"data"`
	TotalCount FlexInt                 `json:"totalCount"`
}

// ComplianceRequirementWithControls represents a compliance requirement with its controls
type ComplianceRequirementWithControls struct {
	RequirementID       string    `json:"requirementId"`
	Name                string    `json:"name"`
	PolicyID            string    `json:"policyId"`
	PolicyName          string    `json:"policyName"`
	Severity            string    `json:"severity"`
	Pass                bool      `json:"pass"`
	FailedControls      int       `json:"failedControls"`
	HighSeverityCount   int       `json:"highSeverityCount"`
	MediumSeverityCount int       `json:"mediumSeverityCount"`
	LowSeverityCount    int       `json:"lowSeverityCount"`
	AcceptedCount       int       `json:"acceptedCount"`
	PassingCount        int       `json:"passingCount"`
	Description         string    `json:"description"`
	Controls            []Control `json:"controls"`
	Zone                Zone      `json:"zone"`
}

// ComplianceResponseWithControls represents the API response with controls included
type ComplianceResponseWithControls struct {
	Data       []ComplianceRequirementWithControls `json:"data"`
	TotalCount FlexInt                             `json:"totalCount"`
}

// Acceptance represents risk acceptance information
type Acceptance struct {
	Justification  string `json:"justification"`
	ExpirationDate string `json:"expirationDate"`
}

// CloudResource represents a resource from Cloud Resources API or Cluster Analysis API
type CloudResource struct {
	Hash         string                 `json:"hash"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Passed       bool                   `json:"passed"`

	// Cloud Resources API フィールド（AWS/GCP/Azure）
	Platform     string                 `json:"platform,omitempty"`
	Account      string                 `json:"account,omitempty"`
	Location     string                 `json:"location,omitempty"`
	Organization string                 `json:"organization,omitempty"`

	// Cluster Analysis API フィールド（Docker/Linux/K8s）
	OSName               string `json:"osName,omitempty"`
	OSImage              string `json:"osImage,omitempty"`
	ClusterName          string `json:"clusterName,omitempty"`
	DistributionName     string `json:"distributionName,omitempty"`
	DistributionVersion  string `json:"distributionVersion,omitempty"`
	PlatformAccountID    string `json:"platformAccountId,omitempty"`
	CloudResourceID      string `json:"cloudResourceId,omitempty"`
	CloudRegion          string `json:"cloudRegion,omitempty"`

	// 共通フィールド
	Acceptance   *Acceptance            `json:"acceptance"`
	Zones        []Zone                 `json:"zones"`
	LastSeenDate string                 `json:"lastSeenDate"`
	LabelValues  []string               `json:"labelValues,omitempty"`
	GlobalID     string                 `json:"globalId,omitempty"`
	AgentTags    []string               `json:"agentTags,omitempty"`
	ConfigError  string                 `json:"configError,omitempty"`
	NodesCount   int                    `json:"nodesCount,omitempty"`

	CreatedAt    time.Time              `json:"createdAt,omitempty"`
	UpdatedAt    time.Time              `json:"updatedAt,omitempty"`
}

// GetAcceptanceStatus returns the acceptance status of the resource
func (r *CloudResource) GetAcceptanceStatus() string {
	if r.Acceptance != nil {
		return "accepted"
	}
	if !r.Passed {
		return "failed"
	}
	return "passed"
}

// CloudResourceResponse represents the API response for cloud resources
type CloudResourceResponse struct {
	Data       []CloudResource `json:"data"`
	TotalCount FlexInt         `json:"totalCount"`
}

// InventoryResponse represents the API response for inventory resources
type InventoryResponse struct {
	Data       []InventoryResource `json:"data"`
	TotalCount FlexInt             `json:"totalCount"`
}

// FlexInt is a custom type that can unmarshal both string and int values
type FlexInt int

// UnmarshalJSON implements custom unmarshaling for FlexInt
func (fi *FlexInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as int first
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*fi = FlexInt(i)
		return nil
	}

	// Try to unmarshal as string
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FlexInt: cannot unmarshal %s into int or string", string(data))
	}

	// Convert string to int
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("FlexInt: cannot convert string %q to int: %w", s, err)
	}

	*fi = FlexInt(i)
	return nil
}

// Int returns the int value of FlexInt
func (fi FlexInt) Int() int {
	return int(fi)
}
