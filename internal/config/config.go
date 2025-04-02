package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type HTTPConfig struct {
	Port               string        `yaml:"port"`
	ReadTimeout        time.Duration `yaml:"readTimeout"`
	WriteTimeout       time.Duration `yaml:"writeTimeout"`
	MaxHeaderBytes     int           `yaml:"maxHeaderBytes"`
	CORSAllowedOrigins []string      `yaml:"corsAllowedOrigins"`
}

type CookiesConfig struct {
	Secure   bool   `yaml:"secure"`
	HTTPOnly bool   `yaml:"httpOnly"`
	SameSite string `yaml:"sameSite"`
}

type CacheConfig struct {
	TTL int `yaml:"ttl"`
}

type PostgresDBConfig struct {
	DSN string
}

type RedisDBConfig struct {
	Address  string
	Password string
	DB       int `yaml:"db"`
	TTL      int `yaml:"ttl"`
}

type Config struct {
	HTTP     HTTPConfig    `yaml:"http"`
	Cookies  CookiesConfig `yaml:"cookies"`
	Cache    CacheConfig   `yaml:"cache"`
	Postgres PostgresDBConfig
	Redis    RedisDBConfig `yaml:"redis"`
	Secrets  struct {
		SessionTTL string `yaml:"session_ttl"`
		CSRF       string `yaml:"csrf"`
	} `yaml:"secrets"`
}

func Load() (*Config, error) {
	// dotenv properties
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("ошибка загрузки .env: %w", err)
	}
	// yml properties
	yamlFile, err := os.ReadFile("configs/main.yml")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	// формирование DSN для PostgreSQL
	cfg.Postgres.DSN = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	// настройки Redis
	cfg.Redis.Address = fmt.Sprintf(
		"%s:%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	)
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	// секреты
	cfg.Secrets.CSRF = os.Getenv("CSRF_SECRET")

	return &cfg, nil
}
