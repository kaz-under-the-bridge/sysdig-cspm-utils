package database

import (
	"database/sql"
	"fmt"

	// SQLite3ドライバーを読み込む
	_ "github.com/mattn/go-sqlite3"
)

const (
	// SQLite schema for CSPM data based on multi-cloud design
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

	createRiskAcceptancesTable = `
	CREATE TABLE IF NOT EXISTS risk_acceptances (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		control_id TEXT NOT NULL,
		description TEXT,
		reason TEXT,
		acceptance_date TEXT NOT NULL,
		username TEXT,
		user_display_name TEXT,
		filter TEXT,
		zone_id TEXT,
		accept_period TEXT,
		expires_at TEXT,
		is_expired BOOLEAN DEFAULT 0,
		is_system BOOLEAN DEFAULT 0,
		type INTEGER,
		source_id TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`

	// Indexes for efficient searching
	createIndexes = `
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
	CREATE INDEX IF NOT EXISTS idx_rel_status ON control_resource_relations(acceptance_status);

	-- Risk Acceptancesのインデックス
	CREATE INDEX IF NOT EXISTS idx_risk_control_id ON risk_acceptances(control_id);
	CREATE INDEX IF NOT EXISTS idx_risk_zone_id ON risk_acceptances(zone_id) WHERE zone_id IS NOT NULL;
	CREATE INDEX IF NOT EXISTS idx_risk_tenant_id ON risk_acceptances(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_risk_is_expired ON risk_acceptances(is_expired);`
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
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// initialize creates the database schema
func (d *Database) initialize() error {
	queries := []string{
		createComplianceRequirementsTable,
		createControlsTable,
		createCloudResourcesTable,
		createControlResourceRelationsTable,
		createRiskAcceptancesTable,
		createIndexes,
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
