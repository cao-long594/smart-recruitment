package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GRPCAddr string
	MySQLDSN string

	OSSBucket    string
	OSSRegion    string
	OSSEndpoint  string
	OSSAccessKey string
	OSSSecretKey string

	QwenAPIKey  string
	QwenModel   string
	QwenBaseURL string
}

type unifiedFile struct {
	Logic struct {
		GRPCAddr string `yaml:"grpc_addr"`
		MySQLDSN string `yaml:"mysql_dsn"`
	} `yaml:"logic"`
	Web struct {
		HTTPAddr      string `yaml:"http_addr"`
		LogicGRPCAddr string `yaml:"logic_grpc_addr"`
		JWTSecret     string `yaml:"jwt_secret"`
	} `yaml:"web"`
	OSS struct {
		Bucket          string `yaml:"bucket"`
		Region          string `yaml:"region"`
		Endpoint        string `yaml:"endpoint"`
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
	} `yaml:"oss"`
	Qwen struct {
		APIKey  string `yaml:"api_key"`
		Model   string `yaml:"model"`
		BaseURL string `yaml:"base_url"`
	} `yaml:"qwen"`
}

// Load 配置优先级：统一 config.yaml → 环境变量（环境变量覆盖同名字段）。
// 配置文件：CONFIG_FILE，或 ./config/config.yaml（请在 logic-grpc-service 目录下启动，或设置 CONFIG_FILE）。
func Load() *Config {
	c := &Config{
		GRPCAddr: ":50051",
		MySQLDSN: "root:root@tcp(127.0.0.1:3306)/recruitment?parseTime=true&charset=utf8mb4",
	}
	applyUnifiedFile(c)
	if c.OSSRegion == "" {
		c.OSSRegion = "us-east-1"
	}
	if c.QwenModel == "" {
		c.QwenModel = "qwen-plus"
	}
	if v := os.Getenv("LOGIC_GRPC_ADDR"); v != "" {
		c.GRPCAddr = v
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		c.MySQLDSN = v
	}
	if v := os.Getenv("OSS_BUCKET"); v != "" {
		c.OSSBucket = v
	}
	if v := os.Getenv("OSS_REGION"); v != "" {
		c.OSSRegion = v
	}
	if v := os.Getenv("OSS_ENDPOINT"); v != "" {
		c.OSSEndpoint = v
	}
	if v := os.Getenv("OSS_ACCESS_KEY_ID"); v != "" {
		c.OSSAccessKey = v
	}
	if v := os.Getenv("OSS_SECRET_ACCESS_KEY"); v != "" {
		c.OSSSecretKey = v
	}
	if v := os.Getenv("QWEN_API_KEY"); v != "" {
		c.QwenAPIKey = v
	}
	if v := os.Getenv("QWEN_MODEL"); v != "" {
		c.QwenModel = v
	}
	if v := os.Getenv("QWEN_BASE_URL"); v != "" {
		c.QwenBaseURL = normalizeQwenBaseURL(v)
	}
	return c
}

func resolveConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("CONFIG_FILE")); p != "" {
		return p
	}
	return "config/config.yaml"
}

func applyUnifiedFile(c *Config) {
	path := resolveConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var f unifiedFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		fmt.Fprintf(os.Stderr, "warn: parse config %s: %v\n", path, err)
		return
	}
	if s := strings.TrimSpace(f.Logic.GRPCAddr); s != "" {
		c.GRPCAddr = s
	}
	if s := strings.TrimSpace(f.Logic.MySQLDSN); s != "" {
		c.MySQLDSN = s
	}
	o := f.OSS
	if strings.TrimSpace(o.Bucket) != "" {
		c.OSSBucket = strings.TrimSpace(o.Bucket)
	}
	if strings.TrimSpace(o.Region) != "" {
		c.OSSRegion = strings.TrimSpace(o.Region)
	}
	if strings.TrimSpace(o.Endpoint) != "" {
		c.OSSEndpoint = strings.TrimSpace(o.Endpoint)
	}
	if strings.TrimSpace(o.AccessKeyID) != "" {
		c.OSSAccessKey = strings.TrimSpace(o.AccessKeyID)
	}
	if strings.TrimSpace(o.SecretAccessKey) != "" {
		c.OSSSecretKey = strings.TrimSpace(o.SecretAccessKey)
	}
	q := f.Qwen
	if s := strings.TrimSpace(q.APIKey); s != "" {
		c.QwenAPIKey = s
	}
	if s := strings.TrimSpace(q.Model); s != "" {
		c.QwenModel = s
	}
	if s := strings.TrimSpace(q.BaseURL); s != "" {
		c.QwenBaseURL = normalizeQwenBaseURL(s)
	}
}

// normalizeQwenBaseURL 将误填的 .../v1/chat/completions 规范为 .../v1（Eino 侧会再拼对话路径）。
func normalizeQwenBaseURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	u = strings.TrimSuffix(u, "/")
	if strings.HasSuffix(u, "/chat/completions") {
		u = strings.TrimSuffix(u, "/chat/completions")
		u = strings.TrimSuffix(u, "/")
	}
	return u
}
