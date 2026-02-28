package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Port != 8080 {
		t.Errorf("Default port should be 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Default host should be 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Storage.DataDir == "" {
		t.Error("DataDir should not be empty")
	}
	if cfg.LLM.Provider != "openai" {
		t.Errorf("Default LLM provider should be openai, got %s", cfg.LLM.Provider)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Default log level should be info, got %s", cfg.Log.Level)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	// 加载不存在的配置文件应该返回默认配置
	cfg, err := Load("/non/existent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Error("Should return default config when file not found")
	}
}

func TestLoad_FromFile(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: 0.0.0.0
  port: 9090
storage:
  data_dir: /custom/data
llm:
  provider: anthropic
  model: claude-3
log:
  level: debug
  format: json
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Host mismatch: got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Port mismatch: got %d", cfg.Server.Port)
	}
	if cfg.Storage.DataDir != "/custom/data" {
		t.Errorf("DataDir mismatch: got %s", cfg.Storage.DataDir)
	}
	if cfg.LLM.Provider != "anthropic" {
		t.Errorf("Provider mismatch: got %s", cfg.LLM.Provider)
	}
	if cfg.LLM.Model != "claude-3" {
		t.Errorf("Model mismatch: got %s", cfg.LLM.Model)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log level mismatch: got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Log format mismatch: got %s", cfg.Log.Format)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("INSIGHT_HUB_DATA_DIR", "/env/override/data")
	defer os.Unsetenv("INSIGHT_HUB_DATA_DIR")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Storage.DataDir != "/env/override/data" {
		t.Errorf("DataDir should be overridden by env: got %s", cfg.Storage.DataDir)
	}
}

func TestLoad_EnvConfigPath(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "custom-config.yaml")

	configContent := `
server:
  port: 7777
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// 设置环境变量指向配置文件
	os.Setenv("INSIGHT_HUB_CONFIG", configPath)
	defer os.Unsetenv("INSIGHT_HUB_CONFIG")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 7777 {
		t.Errorf("Port should be 7777 from env config path: got %d", cfg.Server.Port)
	}
}

func TestExpandPath(t *testing.T) {
	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/data", filepath.Join(homeDir, "data")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~/", homeDir},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if result != tt.expected {
			t.Errorf("expandPath(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 3000,
		},
		Storage: StorageConfig{
			DataDir: "/test/data",
		},
		LLM: LLMConfig{
			Provider: "openai",
			Model:    "gpt-4",
		},
		Log: LogConfig{
			Level:  "debug",
			Format: "json",
		},
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// 重新加载并验证
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Server.Port != 3000 {
		t.Errorf("Port mismatch after reload: got %d", loaded.Server.Port)
	}
	if loaded.LLM.Model != "gpt-4" {
		t.Errorf("Model mismatch after reload: got %s", loaded.LLM.Model)
	}
}

func TestConfig_Save_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nested", "dir", "config.yaml")

	cfg := DefaultConfig()
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created in nested directory")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	invalidYAML := `
server:
  port: not-a-number
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}
