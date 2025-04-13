package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type HTTPConfig struct {
	Host               string        `yaml:"host"`
	Port               string        `yaml:"port"`
	ReadTimeout        time.Duration `yaml:"readTimeout"`
	WriteTimeout       time.Duration `yaml:"writeTimeout"`
	MaxHeaderBytes     int           `yaml:"maxHeaderBytes"`
	CORSAllowedOrigins []string      `yaml:"corsAllowedOrigins"`
}

type SessionConfig struct {
	CookieName string        `yaml:"cookieName"`
	Lifetime   time.Duration `yaml:"lifetime"`
	HttpOnly   bool          `yaml:"httpOnly"`
	Secure     bool          `yaml:"secure"`
	SameSite   string        `yaml:"sameSite"`
	Secret     string        `yaml:"-"`
}

type CSRFConfig struct {
	CookieName string        `yaml:"cookieName"`
	Lifetime   time.Duration `yaml:"lifetime"`
	Secret     string        `yaml:"-"`
	HttpOnly   bool          `yaml:"httpOnly"`
	Secure     bool          `yaml:"secure"`
	SameSite   string        `yaml:"sameSite"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"-"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	DSN      string `yaml:"-"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"-"`
	DB       int    `yaml:"db"`
	TTL      int    `yaml:"ttl"`
}

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Session  SessionConfig  `yaml:"session_id"`
	CSRF     CSRFConfig     `yaml:"csrf"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

func Load() (*Config, error) {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	// Чтение YAML
	yamlFile, err := os.ReadFile("configs/main.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	// Заполнение секретов из .env
	cfg.CSRF.Secret = os.Getenv("CSRF_SECRET")

	// Формирование DSN для PostgreSQL
	cfg.Postgres = PostgresConfig{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_CONTAINER_PORT"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		DBName:   os.Getenv("POSTGRES_DB"),
		SSLMode:  "disable",
		DSN: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("POSTGRES_HOST"),
			os.Getenv("POSTGRES_CONTAINER_PORT"),
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
		),
	}

	// Настройка Redis
	cfg.Redis = RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_CONTAINER_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       cfg.Redis.DB,
		TTL:      cfg.Redis.TTL,
	}

	return &cfg, nil
}
