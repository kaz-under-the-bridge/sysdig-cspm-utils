package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// SQLite schema for CSPM data based on multi-cloud design
	createInventoryResourcesTable = `
	CREATE TABLE IF NOT EXISTS inventory_resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		platform TEXT NOT NULL,  -- 'AWS', 'GCP', 'Azure', 'Kubernetes'
		category TEXT,

		-- 共通的な重要メタデータ（NULL許可）
		organization TEXT,    -- AWS: organization, GCP: organization ID
		region TEXT,          -- AWS/GCP: region, Azure: location

		-- プラットフォーム別カラム（NULL許可）
		aws_account TEXT,     -- AWS専用
		aws_arn TEXT,         -- AWS専用
		gcp_project TEXT,     -- GCP専用
		gcp_resource_id TEXT, -- GCP専用
		azure_subscription TEXT, -- Azure専用
		azure_resource_id TEXT,  -- Azure専用
		k8s_namespace TEXT,   -- Kubernetes専用
		k8s_cluster TEXT,     -- Kubernetes専用

		-- 完全なmetadataはJSON保存（拡張性確保）
		metadata_json TEXT,  -- 完全なmetadataオブジェクト
		labels_json TEXT,    -- ラベル配列
		zones_json TEXT,     -- ゾーン配列

		-- ポリシー関連情報もJSON
		posture_control_summary_json TEXT,
		posture_policy_summary_json TEXT,

		-- その他
		config_api_endpoint TEXT,
		resource_origin TEXT,
		last_seen BIGINT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	createComplianceRequirementsTable = `
	CREATE TABLE IF NOT EXISTS compliance_requirements (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		requirement_id TEXT NOT NULL,
		name TEXT NOT NULL,
		policy_id TEXT NOT NULL,
		policy_name TEXT NOT NULL,
		policy_type TEXT,  -- 'CIS', 'SOC2', 'PCI-DSS', 'HIPAA', etc.
		platform TEXT,     -- 'AWS', 'GCP', 'Azure', 'Multi-Cloud'
		severity TEXT NOT NULL,
		pass BOOLEAN NOT NULL,
		zone_id TEXT,
		zone_name TEXT,
		failed_controls INTEGER DEFAULT 0,
		high_severity_count INTEGER DEFAULT 0,
		medium_severity_count INTEGER DEFAULT 0,
		low_severity_count INTEGER DEFAULT 0,
		accepted_count INTEGER DEFAULT 0,
		passing_count INTEGER DEFAULT 0,
		description TEXT,
		resource_api_endpoint TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(requirement_id, policy_id, zone_id)
	)`

	createControlsTable = `
	CREATE TABLE IF NOT EXISTS controls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		control_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		description TEXT,
		requirement_id TEXT NOT NULL,
		severity TEXT NOT NULL,
		pass BOOLEAN NOT NULL,
		objects_count INTEGER DEFAULT 0,      -- Failed数
		passing_count INTEGER DEFAULT 0,      -- Passed数
		accepted_count INTEGER DEFAULT 0,     -- Accepted数
		resource_kind TEXT,
		resource_api_endpoint TEXT NOT NULL,
		target TEXT,
		platform TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (requirement_id) REFERENCES compliance_requirements(requirement_id)
	)`

	createCloudResourcesTable = `
	CREATE TABLE IF NOT EXISTS cloud_resources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		type TEXT NOT NULL,

		-- Cloud Resources API用フィールド（AWS/GCP/Azure）
		platform TEXT,                        -- AWS, GCP, Azure
		account TEXT,                         -- AWS: account名, GCP: project ID, Azure: subscription
		location TEXT,                        -- AWS: region, GCP: region, Azure: location
		organization TEXT,                    -- AWS: o-xxxxx, GCP: 組織ID, Azure: tenant ID

		-- Cluster Analysis API用フィールド（Docker/Linux/Kubernetes）
		os_name TEXT,                         -- linux, windows等
		os_image TEXT,                        -- Amazon Linux 2023等
		cluster_name TEXT,                    -- クラスタ名
		distribution_name TEXT,               -- Linux等
		distribution_version TEXT,            -- ディストリビューションバージョン

		-- 共通メタデータ（JSON形式）
		zones_json TEXT,
		label_values_json TEXT,
		additional_metadata_json TEXT,        -- プラットフォーム固有の追加情報
		last_seen_date TEXT,
		global_id TEXT,

		-- Cluster Analysis API追加フィールド
		platform_account_id TEXT,
		cloud_resource_id TEXT,
		cloud_region TEXT,
		agent_tags_json TEXT,

		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	createControlResourceRelationsTable = `
	CREATE TABLE IF NOT EXISTS control_resource_relations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		control_id TEXT NOT NULL,
		resource_hash TEXT NOT NULL,
		passed BOOLEAN NOT NULL,
		acceptance_status TEXT NOT NULL,  -- 'failed', 'passed', 'accepted'
		acceptance_justification TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

		FOREIGN KEY (control_id) REFERENCES controls(control_id),
		FOREIGN KEY (resource_hash) REFERENCES cloud_resources(hash),
		UNIQUE(control_id, resource_hash)
	)`

	// Indexes for efficient searching
	createIndexes = `
	-- 検索用インデックス（共通）
	CREATE INDEX IF NOT EXISTS idx_res_platform ON inventory_resources(platform);
	CREATE INDEX IF NOT EXISTS idx_res_type ON inventory_resources(type);
	CREATE INDEX IF NOT EXISTS idx_res_organization ON inventory_resources(organization);
	CREATE INDEX IF NOT EXISTS idx_res_region ON inventory_resources(region);
	CREATE INDEX IF NOT EXISTS idx_res_category ON inventory_resources(category);

	-- プラットフォーム別インデックス
	CREATE INDEX IF NOT EXISTS idx_res_aws_account ON inventory_resources(aws_account) WHERE aws_account IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_res_aws_arn ON inventory_resources(aws_arn) WHERE aws_arn IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_res_gcp_project ON inventory_resources(gcp_project) WHERE gcp_project IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_res_azure_subscription ON inventory_resources(azure_subscription) WHERE azure_subscription IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_res_k8s_namespace ON inventory_resources(k8s_namespace) WHERE k8s_namespace IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_res_k8s_cluster ON inventory_resources(k8s_cluster) WHERE k8s_cluster IS NOT NULL;

	-- コンプライアンス要件のインデックス
	CREATE INDEX IF NOT EXISTS idx_req_policy_type ON compliance_requirements(policy_type);
	CREATE INDEX IF NOT EXISTS idx_req_platform ON compliance_requirements(platform);

	-- コントロールのインデックス
	CREATE INDEX IF NOT EXISTS idx_ctrl_requirement ON controls(requirement_id);
	CREATE INDEX IF NOT EXISTS idx_ctrl_id ON controls(control_id);

	-- Cloud Resourcesのインデックス
	CREATE INDEX IF NOT EXISTS idx_cloud_res_hash ON cloud_resources(hash);
	CREATE INDEX IF NOT EXISTS idx_cloud_res_type ON cloud_resources(type);
	CREATE INDEX IF NOT EXISTS idx_cloud_res_platform ON cloud_resources(platform) WHERE platform IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_cloud_res_account ON cloud_resources(account) WHERE account IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_cloud_res_location ON cloud_resources(location) WHERE location IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_cloud_res_os_name ON cloud_resources(os_name) WHERE os_name IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_cloud_res_cluster_name ON cloud_resources(cluster_name) WHERE cluster_name IS NOT NULL;

	-- Control-Resource関連のインデックス
	CREATE INDEX IF NOT EXISTS idx_rel_control ON control_resource_relations(control_id);
	CREATE INDEX IF NOT EXISTS idx_rel_resource ON control_resource_relations(resource_hash);
	CREATE INDEX IF NOT EXISTS idx_rel_status ON control_resource_relations(acceptance_status);`

	// Views for platform-specific analysis
	createAWSResourcesView = `
	CREATE VIEW IF NOT EXISTS aws_resources AS
	SELECT
		hash,
		name,
		type,
		aws_account as account,
		aws_arn as arn,
		region,
		organization,
		json_extract(metadata_json, '$.tags') as tags,
		json_extract(posture_policy_summary_json, '$.passPercentage') as compliance_rate
	FROM inventory_resources
	WHERE platform = 'AWS'`

	createGCPResourcesView = `
	CREATE VIEW IF NOT EXISTS gcp_resources AS
	SELECT
		hash,
		name,
		type,
		gcp_project as project,
		gcp_resource_id as resource_id,
		region,
		organization,
		json_extract(metadata_json, '$.labels') as labels,
		json_extract(posture_policy_summary_json, '$.passPercentage') as compliance_rate
	FROM inventory_resources
	WHERE platform = 'GCP'`

	createMultiCloudSummaryView = `
	CREATE VIEW IF NOT EXISTS multi_cloud_summary AS
	SELECT
		platform,
		CASE
			WHEN platform = 'AWS' THEN aws_account
			WHEN platform = 'GCP' THEN gcp_project
			WHEN platform = 'Azure' THEN azure_subscription
			ELSE 'N/A'
		END as account_identifier,
		organization,
		region,
		COUNT(*) as resource_count,
		COUNT(DISTINCT type) as resource_types,
		AVG(CASE
			WHEN json_extract(posture_policy_summary_json, '$.passPercentage') IS NOT NULL
			THEN json_extract(posture_policy_summary_json, '$.passPercentage')
			ELSE 0
		END) as avg_compliance_rate
	FROM inventory_resources
	GROUP BY platform, account_identifier, organization, region`

	createResourceMetadataExpandedView = `
	CREATE VIEW IF NOT EXISTS resource_metadata_expanded AS
	SELECT
		hash,
		name,
		platform,
		type,
		-- 共通フィールド
		organization,
		region,
		-- プラットフォーム別の主要識別子
		CASE platform
			WHEN 'AWS' THEN aws_account
			WHEN 'GCP' THEN gcp_project
			WHEN 'Azure' THEN azure_subscription
			WHEN 'Kubernetes' THEN k8s_cluster
		END as primary_scope,
		-- JSONから追加メタデータを抽出
		json_extract(metadata_json, '$.tags') as tags,
		json_extract(metadata_json, '$.labels') as labels,
		json_extract(metadata_json, '$.createdTime') as created_time,
		-- コンプライアンス情報
		json_extract(posture_policy_summary_json, '$.passPercentage') as compliance_rate,
		json_extract(posture_control_summary_json, '$[0].failedControls') as primary_failed_controls
	FROM inventory_resources`
)

// Database represents a SQLite database connection with CSPM schema
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection and initializes the schema
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	if err := database.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// initialize creates the database schema
func (d *Database) initialize() error {
	queries := []string{
		createInventoryResourcesTable,
		createComplianceRequirementsTable,
		createControlsTable,
		createCloudResourcesTable,
		createControlResourceRelationsTable,
		createIndexes,
		createAWSResourcesView,
		createGCPResourcesView,
		createMultiCloudSummaryView,
		createResourceMetadataExpandedView,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute schema query: %w", err)
		}
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// DB returns the underlying sql.DB instance
func (d *Database) DB() *sql.DB {
	return d.db
}
