package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/webcore-go/webcore/app/helper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	PubSub   PubSubConfig   `mapstructure:"pubsub"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type AppConfig struct {
	Name            string          `mapstructure:"name"`
	Version         string          `mapstructure:"version"`
	Environment     string          `mapstructure:"environment"`
	Features        FeaturesConfig  `mapstructure:"features"`
	Logging         LoggingConfig   `mapstructure:"logging"`
	CORS            CORSConfig      `mapstructure:"cors"`
	RateLimit       RateLimitConfig `mapstructure:"rate_limit"`
	SecurityHeaders bool            `mapstructure:"security_headers"`
	Module          ModuleConfig    `mapstructure:"module"`
}

type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Max     int  `mapstructure:"max"`
}

type FeaturesConfig struct {
	Recovery  bool `mapstructure:"recovery"`
	Metrics   bool `mapstructure:"metrics"`
	Tracing   bool `mapstructure:"tracing"`
	Profiling bool `mapstructure:"profiling"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	PathPrefix   string        `mapstructure:"path"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver          string            `mapstructure:"driver"` // supported: "postgres", "mysql", "sqlite", "mongodb"
	Host            string            `mapstructure:"host"`
	Port            int               `mapstructure:"port"`
	User            string            `mapstructure:"user"`
	Password        string            `mapstructure:"password"`
	Name            string            `mapstructure:"name"`
	SSLMode         string            `mapstructure:"ssl_mode"`
	Attributes      map[string]string `mapstructure:"attributes"` // Additional connection parameters
	MaxIdleConns    int               `mapstructure:"max_idle_conns"`
	MaxOpenConns    int               `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration     `mapstructure:"conn_max_lifetime"`
	SlaveHosts      []DatabaseConfig  `mapstructure:"slave_hosts"`
}

type RedisConfig struct {
	Host       string        `mapstructure:"host"`
	Port       int           `mapstructure:"port"`
	Password   string        `mapstructure:"password"`
	DB         int           `mapstructure:"db"`
	SlaveHosts []RedisConfig `mapstructure:"slave_hosts"`
}

type KafkaConfig struct {
	Brokers         []string `mapstructure:"brokers"`
	GroupID         string   `mapstructure:"group_id"`
	Topics          []string `mapstructure:"topics"`
	AutoOffsetReset string   `mapstructure:"offset_reset"`
}

type PubSubConfig struct {
	ProjectID       string         `mapstructure:"project_id"`
	Subscription    string         `mapstructure:"subscription"`
	Topic           string         `mapstructure:"topic"`
	CredentialsPath string         `mapstructure:"credentials"`
	Consumer        ConsumerConfig `mapstructure:"consumer"`
	Producer        ProducerConfig `mapstructure:"producer"`
}

type ConsumerConfig struct {
	MaxMessagesPerPull    int               `mapstructure:"maxmessages"`
	SleepTimeBetweenPulls time.Duration     `mapstructure:"sleeptime"`
	AcknowledgeTimeout    time.Duration     `mapstructure:"acktimeout"`
	RetryCount            int               `mapstructure:"retrycount"`
	RetryDelay            time.Duration     `mapstructure:"retrydelay"`
	FlowControl           FlowControlConfig `mapstructure:"flowcontrol"`
}

type FlowControlConfig struct {
	Enabled                bool  `mapstructure:"enabled"`
	MaxOutstandingMessages int   `mapstructure:"maxmessages"`
	MaxOutstandingBytes    int64 `mapstructure:"maxbytes"`
}

type ProducerConfig struct {
	EnableMessageOrdering bool              `mapstructure:"enableordering"`
	BatchSize             int               `mapstructure:"batchsize"`
	MessageAttributes     map[string]string `mapstructure:"attributes"`
}

type AuthConfig struct {
	Control      string        `mapstructure:"control"` // e.g., "RBAC", "ABAC"
	Store        string        `mapstructure:"store"`   // e.g., "yaml", "db"
	Type         string        `mapstructure:"type"`    // e.g., "jwt", "apikey"
	SecretKey    string        `mapstructure:"secret_key"`
	ExpiresIn    time.Duration `mapstructure:"expires_in"`     // In seconds
	APIKeyHeader string        `mapstructure:"api_key_header"` // Header name for API key (default: "X-API-Key")
	APIKeyPrefix string        `mapstructure:"api_key_prefix"` // Optional prefix for API key validation
}

type ModuleConfig struct {
	Disabled []string `mapstructure:"disabled"`
	BasePath string   `mapstructure:"base_path"`
}

func (c *Config) GetFiberConfig(errorHandler fiber.ErrorHandler) fiber.Config {
	return fiber.Config{
		ReadTimeout:   c.Server.ReadTimeout,
		WriteTimeout:  c.Server.WriteTimeout,
		CaseSensitive: true,
		StrictRouting: true,
		ErrorHandler:  errorHandler,

		// For faster encoder/decoder use go-json that wrap in helper.JSONMarshal and helper.JSONUnmarshal
		JSONEncoder: helper.JSONMarshal,
		JSONDecoder: helper.JSONUnmarshal,
	}
}
