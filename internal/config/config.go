package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Bifrost  BifrostConfig
	Cost     CostConfig
	Log      LogConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string
	Port int
}

func (s *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// BifrostConfig Bifrost 配置
type BifrostConfig struct {
	ConfigPath string
}

// CostConfig 成本优化配置
type CostConfig struct {
	DefaultProfitMargin float64
	MinProfitMargin     float64
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string
	Format string
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// 如果 .env 文件不存在，使用环境变量
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "llm_gateway"),
			Password: getEnv("DB_PASSWORD", "llm_gateway"),
			DBName:   getEnv("DB_NAME", "llm_gateway"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Bifrost: BifrostConfig{
			ConfigPath: getEnv("BIFROST_CONFIG_PATH", "./config/bifrost.json"),
		},
		Cost: CostConfig{
			DefaultProfitMargin: getEnvFloat("DEFAULT_PROFIT_MARGIN", 0.1),
			MinProfitMargin:     getEnvFloat("MIN_PROFIT_MARGIN", 0.05),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "debug"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}
