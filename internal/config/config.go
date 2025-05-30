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

type MinioConfig struct {
	InternalEndpoint string `yaml:"internal_endpoint"`
	PublicEndpoint   string `yaml:"public_endpoint"`
	RootUser         string `yaml:"-"`
	RootPassword     string `yaml:"-"`
	UseSSL           bool   `yaml:"use_ssl"`
	Scheme           string `yaml:"scheme"`
	Bucket           string `yaml:"bucket"`
}

type S3ClientConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (s3 *S3ClientConfig) Addr() string {
	return fmt.Sprintf("%s:%s", s3.Host, s3.Port)
}

type S3Config struct {
	Host       string         `yaml:"host"`
	Port       string         `yaml:"port"`
	MetricPort string         `yaml:"metric_port"`
	Minio      MinioConfig    `yaml:"minio"`
	Postgres   PostgresConfig `yaml:"postgres"`
}

func (s3 *S3Config) Addr() string {
	return fmt.Sprintf("%s:%s", s3.Host, s3.Port)
}

type RedisConfig struct {
	Host     string       `yaml:"host"`
	Port     string       `yaml:"port"`
	Password string       `yaml:"-"`
	DB       int          `yaml:"db"`
	TTL      int          `yaml:"ttl"`
	Pool     RedisPoolCfg `yaml:"pool"`
}

type RedisPoolCfg struct {
	MaxIdle     int           `yaml:"maxIdle"`
	MaxActive   int           `yaml:"maxActive"`
	IdleTimeout time.Duration `yaml:"idleTimeout"`
}

type MicroservicesConfig struct {
	Auth AuthClientConfig `yaml:"auth_service"`
	S3   S3ClientConfig   `yaml:"static_service"`
}

type AuthConfig struct {
	Host       string      `yaml:"host"`
	Port       string      `yaml:"port"`
	MetricPort string      `yaml:"metric_port"`
	Redis      RedisConfig `yaml:"redis"`
}

func (a *AuthConfig) Addr() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

type AuthClientConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (a *AuthClientConfig) Addr() string {
	return fmt.Sprintf("%s:%s", a.Host, a.Port)
}

type ResumeConfig struct {
	StaticPath  string `yaml:"staticPath"`
	StaticFile  string `yaml:"staticFile"`
	PaperWidth  string `yaml:"paperWidth"`
	PaperHeight string `yaml:"paperHeight"`
	GenerateURL string `yaml:"generateURL"`
}

type Config struct {
	HTTP          HTTPConfig          `yaml:"http"`
	Session       SessionConfig       `yaml:"session_id"`
	CSRF          CSRFConfig          `yaml:"csrf"`
	Postgres      PostgresConfig      `yaml:"postgres"`
	Microservices MicroservicesConfig `yaml:"microservices"`
	Resume        ResumeConfig        `yaml:"resume"`
}

func LoadAppConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	yamlFile, err := os.ReadFile("configs/main.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing YAML: %w", err)
	}

	cfg.CSRF.Secret = os.Getenv("CSRF_SECRET")

	cfg.Postgres = loadPostgresConfig()

	return &cfg, nil
}

func LoadAuthConfig() (*AuthConfig, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("ошибка загрузки .env: %w", err)
	}

	yamlFile, err := os.ReadFile("configs/auth.yml")
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла конфигурации сервиса авторизации: %w", err)
	}

	var cfg AuthConfig
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	cfg.Redis = RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_CONTAINER_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       cfg.Redis.DB,
		TTL:      cfg.Redis.TTL,
		//Pool:     pool,
	}
	return &cfg, nil
}

func LoadS3Config() (*S3Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env: %w", err)
	}

	yamlFile, err := os.ReadFile("configs/static.yml")
	if err != nil {
		return nil, fmt.Errorf("error reading static config file: %w", err)
	}

	var cfg S3Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing static YAML: %w", err)
	}

	cfg.Minio.RootUser = os.Getenv("MINIO_ROOT_USER")
	cfg.Minio.RootPassword = os.Getenv("MINIO_ROOT_PASSWORD")

	cfg.Postgres = loadPostgresConfig()

	return &cfg, nil
}

func loadPostgresConfig() PostgresConfig {
	return PostgresConfig{
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
}
