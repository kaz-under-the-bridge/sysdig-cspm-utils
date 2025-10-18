package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromEnvironment(t *testing.T) {
	// 環境変数を設定
	os.Setenv("SYSDIG_API_TOKEN", "test-token")
	os.Setenv("SYSDIG_API_URL", "https://test.sysdig.com")
	defer func() {
		os.Unsetenv("SYSDIG_API_TOKEN")
		os.Unsetenv("SYSDIG_API_URL")
	}()

	cfg, err := Load("", "", "")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.APIToken != "test-token" {
		t.Errorf("Expected APIToken 'test-token', got '%s'", cfg.APIToken)
	}
	if cfg.APIURL != "https://test.sysdig.com" {
		t.Errorf("Expected APIURL 'https://test.sysdig.com', got '%s'", cfg.APIURL)
	}
}

func TestLoad_FromFile(t *testing.T) {
	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	cfg := &Config{
		APIToken: "file-token",
		APIURL:   "https://file.sysdig.com",
	}

	err := cfg.Save(configFile)
	if err != nil {
		t.Fatalf("Failed to save config file: %v", err)
	}

	// ファイルから読み込み
	loadedCfg, err := Load(configFile, "", "")
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if loadedCfg.APIToken != "file-token" {
		t.Errorf("Expected APIToken 'file-token', got '%s'", loadedCfg.APIToken)
	}
	if loadedCfg.APIURL != "https://file.sysdig.com" {
		t.Errorf("Expected APIURL 'https://file.sysdig.com', got '%s'", loadedCfg.APIURL)
	}
}

func TestLoad_CommandLineOverride(t *testing.T) {
	// 環境変数を設定
	os.Setenv("SYSDIG_API_TOKEN", "env-token")
	os.Setenv("SYSDIG_API_URL", "https://env.sysdig.com")
	defer func() {
		os.Unsetenv("SYSDIG_API_TOKEN")
		os.Unsetenv("SYSDIG_API_URL")
	}()

	// コマンドラインフラグで上書き
	cfg, err := Load("", "cli-token", "https://cli.sysdig.com")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.APIToken != "cli-token" {
		t.Errorf("Expected APIToken 'cli-token', got '%s'", cfg.APIToken)
	}
	if cfg.APIURL != "https://cli.sysdig.com" {
		t.Errorf("Expected APIURL 'https://cli.sysdig.com', got '%s'", cfg.APIURL)
	}
}

func TestLoad_Priority(t *testing.T) {
	// 環境変数を設定
	os.Setenv("SYSDIG_API_TOKEN", "env-token")
	defer os.Unsetenv("SYSDIG_API_TOKEN")

	// テスト用の設定ファイルを作成
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")
	fileCfg := &Config{
		APIToken: "file-token",
		APIURL:   "https://file.sysdig.com",
	}
	err := fileCfg.Save(configFile)
	if err != nil {
		t.Fatalf("Failed to save config file: %v", err)
	}

	// 優先順位: CLI > File > Environment
	// ここではCLIでトークンのみ指定、URLはファイルから
	cfg, err := Load(configFile, "cli-token", "")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// CLIで指定したトークンが優先
	if cfg.APIToken != "cli-token" {
		t.Errorf("Expected APIToken 'cli-token', got '%s'", cfg.APIToken)
	}
	// URLはファイルから
	if cfg.APIURL != "https://file.sysdig.com" {
		t.Errorf("Expected APIURL 'https://file.sysdig.com', got '%s'", cfg.APIURL)
	}
}

func TestLoad_DefaultURL(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("SYSDIG_API_TOKEN")
	os.Unsetenv("SYSDIG_API_URL")

	cfg, err := Load("", "", "")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// デフォルトURLが設定されているか確認
	if cfg.APIURL != "https://us2.app.sysdig.com" {
		t.Errorf("Expected default APIURL 'https://us2.app.sysdig.com', got '%s'", cfg.APIURL)
	}
}

func TestSave(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	cfg := &Config{
		APIToken: "save-test-token",
		APIURL:   "https://save.sysdig.com",
	}

	err := cfg.Save(configFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// ファイルが作成されたか確認
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Errorf("Config file was not created: %s", configFile)
	}

	// 保存したファイルを読み込んで確認
	loadedCfg, err := Load(configFile, "", "")
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedCfg.APIToken != cfg.APIToken {
		t.Errorf("Expected APIToken '%s', got '%s'", cfg.APIToken, loadedCfg.APIToken)
	}
	if loadedCfg.APIURL != cfg.APIURL {
		t.Errorf("Expected APIURL '%s', got '%s'", cfg.APIURL, loadedCfg.APIURL)
	}
}

func TestLoadFromFile_NotFound(t *testing.T) {
	_, err := Load("/nonexistent/config.json", "", "")
	if err == nil {
		t.Error("Expected error for non-existent config file, got nil")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid.json")

	// 無効なJSONファイルを作成
	err := os.WriteFile(configFile, []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	_, err = Load(configFile, "", "")
	if err == nil {
		t.Error("Expected error for invalid JSON config file, got nil")
	}
}
