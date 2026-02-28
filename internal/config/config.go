package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	LLM      LLMConfig      `yaml:"llm"`
	Log      LogConfig      `yaml:"log"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	Name            string `yaml:"name"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	SSLMode         string `yaml:"sslmode"`
	MaxConns        int32  `yaml:"max_conns"`
	MinConns        int32  `yaml:"min_conns"`
	MaxConnLifetime string `yaml:"max_conn_lifetime"`
	MaxConnIdleTime string `yaml:"max_conn_idle_time"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	DataDir string `yaml:"data_dir"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider string `yaml:"provider"` // openai, anthropic, local
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	BaseURL  string `yaml:"base_url"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`   // debug, info, warn, error
	Format string `yaml:"format"`  // json, text
	Output string `yaml:"output"`  // stdout, stderr, file path
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8090,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5433,
			Name:            "insight_hub",
			User:            "insight",
			Password:        "insight_local_dev",
			SSLMode:         "disable",
			MaxConns:        10,
			MinConns:        2,
			MaxConnLifetime: "1h",
			MaxConnIdleTime: "10m",
		},
		Storage: StorageConfig{
			DataDir: filepath.Join(homeDir, ".insight-hub"),
		},
		LLM: LLMConfig{
			Provider: "openai",
			Model:    "gpt-4o-mini",
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// Load 加载配置
// 优先级：configPath > INSIGHT_HUB_CONFIG > ~/.insight-hub/config.yaml
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	// 确定配置文件路径
	if configPath == "" {
		configPath = os.Getenv("INSIGHT_HUB_CONFIG")
	}
	if configPath == "" {
		homeDir, _ := os.UserHomeDir()
		configPath = filepath.Join(homeDir, ".insight-hub", "config.yaml")
	}

	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 配置文件不存在，使用默认配置
		return cfg, nil
	}

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析 YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// 环境变量覆盖
	if dataDir := os.Getenv("INSIGHT_HUB_DATA_DIR"); dataDir != "" {
		cfg.Storage.DataDir = dataDir
	}

	// Docker 环境变量覆盖数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}

	// Docker 环境变量覆盖服务器配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}

	// 展开 ~ 路径
	cfg.Storage.DataDir = expandPath(cfg.Storage.DataDir)

	return cfg, nil
}

// expandPath 展开 ~ 为用户主目录
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, strings.TrimPrefix(path, "~/"))
	}
	return path
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// 序列化为 YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
