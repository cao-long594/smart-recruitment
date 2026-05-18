package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTPAddr      string
	LogicGRPCAddr string
	JWTSecret     string
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

// Load 配置优先级：与 logic 服务相同的统一 config.yaml → 环境变量覆盖。
// 配置文件：CONFIG_FILE，或 ./config/config.yaml、../logic-grpc-service/config/config.yaml（在 web-gin-service 目录下启动时）。
func Load() *Config {
	c := &Config{
		HTTPAddr:      ":8080",
		LogicGRPCAddr: "127.0.0.1:50051",
		JWTSecret:     "dev-change-me-in-production",
	}
	applyUnifiedFile(c)
	if v := os.Getenv("WEB_HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("LOGIC_GRPC_ADDR"); v != "" {
		c.LogicGRPCAddr = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	return c
}

func resolveConfigPath() string {
	if p := strings.TrimSpace(os.Getenv("CONFIG_FILE")); p != "" {
		return p
	}
	for _, rel := range []string{"config/config.yaml", "../logic-grpc-service/config/config.yaml"} {
		if st, err := os.Stat(rel); err == nil && !st.IsDir() {
			return rel
		}
	}
	return "../logic-grpc-service/config/config.yaml"
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
	w := f.Web
	if strings.TrimSpace(w.HTTPAddr) != "" {
		c.HTTPAddr = strings.TrimSpace(w.HTTPAddr)
	}
	if strings.TrimSpace(w.LogicGRPCAddr) != "" {
		c.LogicGRPCAddr = strings.TrimSpace(w.LogicGRPCAddr)
	}
	if strings.TrimSpace(w.JWTSecret) != "" {
		c.JWTSecret = strings.TrimSpace(w.JWTSecret)
	}
}
