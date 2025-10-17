package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type ServiceConfig struct {
	ServiceName string            `mapstructure:"service_name"`
	Port        string            `mapstructure:"port"`
	Environment string            `mapstructure:"environment"`
	LogLevel    string            `mapstructure:"log_level"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Services    ServicesConfig    `mapstructure:"services"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret         string        `mapstructure:"secret"`
	Expiration     time.Duration `mapstructure:"expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
}

type ServicesConfig struct {
	UserService    ServiceEndpoint `mapstructure:"user_service"`
	ProductService ServiceEndpoint `mapstructure:"product_service"`
	OrderService   ServiceEndpoint `mapstructure:"order_service"`
	GatewayService ServiceEndpoint `mapstructure:"gateway_service"`
}

type ServiceEndpoint struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	URL  string `mapstructure:"url"`
}

func (se ServiceEndpoint) GetServiceURL() string {
	if se.URL != "" {
		return se.URL
	}
	return fmt.Sprintf("http://%s:%s", se.Host, se.Port)
}

type MonitoringConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Grafana    GrafanaConfig    `mapstructure:"grafana"`
	Loki       LokiConfig       `mapstructure:"loki"`
}

type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    string `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type GrafanaConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
	User    string `mapstructure:"user"`
	Pass    string `mapstructure:"pass"`
}

type LokiConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	URL     string `mapstructure:"url"`
}

func LoadConfig(serviceName string) (*ServiceConfig, error) {
	// Get environment from ENV variable, default to development
	environment := getEnvironment()
	
	// Set config name based on environment
	configName := fmt.Sprintf("config.%s", environment)
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	
	// Add config paths in order of precedence
	viper.AddConfigPath(fmt.Sprintf("./services/%s/configs", serviceName)) // Service-specific config
	viper.AddConfigPath("./deployments/configs")                          // Global config
	viper.AddConfigPath("./configs")                                      // Fallback config
	viper.AddConfigPath(".")                                              // Current directory

	// Enable automatic environment variable binding
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	bindEnvVariables(serviceName)

	// Set defaults
	setDefaults(serviceName)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// No fallback to hardcoded config files for security
		// Configuration must be generated from templates
		return nil, fmt.Errorf("configuration file not found for environment '%s'. Please run 'make config-%s' to generate it", environment, environment)
	}

	// Validate required configuration
	if err := validateConfig(serviceName); err != nil {
		return nil, err
	}

	// Unmarshal into struct
	var config ServiceConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Set service name and environment
	config.ServiceName = serviceName
	config.Environment = environment

	return &config, nil
}

// bindEnvVariables binds environment variables to viper keys
func bindEnvVariables(serviceName string) {
	// Service configuration
	viper.BindEnv("service_name", "SERVICE_NAME")
	viper.BindEnv("port", "PORT")
	viper.BindEnv("environment", "ENVIRONMENT")
	viper.BindEnv("log_level", "LOG_LEVEL")

	// Database configuration
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")

	// Redis configuration
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// JWT configuration
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expiration", "JWT_EXPIRATION")
	viper.BindEnv("jwt.refresh_expiration", "JWT_REFRESH_EXPIRATION")

	// Service endpoints
	viper.BindEnv("services.user_service.host", "USER_SERVICE_HOST")
	viper.BindEnv("services.user_service.port", "USER_SERVICE_PORT")
	viper.BindEnv("services.user_service.url", "USER_SERVICE_URL")
	
	viper.BindEnv("services.product_service.host", "PRODUCT_SERVICE_HOST")
	viper.BindEnv("services.product_service.port", "PRODUCT_SERVICE_PORT")
	viper.BindEnv("services.product_service.url", "PRODUCT_SERVICE_URL")
	
	viper.BindEnv("services.order_service.host", "ORDER_SERVICE_HOST")
	viper.BindEnv("services.order_service.port", "ORDER_SERVICE_PORT")
	viper.BindEnv("services.order_service.url", "ORDER_SERVICE_URL")
	
	viper.BindEnv("services.gateway_service.host", "GATEWAY_SERVICE_HOST")
	viper.BindEnv("services.gateway_service.port", "GATEWAY_SERVICE_PORT")
	viper.BindEnv("services.gateway_service.url", "GATEWAY_SERVICE_URL")

	// Monitoring configuration
	viper.BindEnv("monitoring.prometheus.enabled", "PROMETHEUS_ENABLED")
	viper.BindEnv("monitoring.prometheus.port", "PROMETHEUS_PORT")
	viper.BindEnv("monitoring.prometheus.path", "PROMETHEUS_PATH")
	
	viper.BindEnv("monitoring.grafana.enabled", "GRAFANA_ENABLED")
	viper.BindEnv("monitoring.grafana.url", "GRAFANA_URL")
	viper.BindEnv("monitoring.grafana.user", "GRAFANA_USER")
	viper.BindEnv("monitoring.grafana.pass", "GRAFANA_PASS")
	
	viper.BindEnv("monitoring.loki.enabled", "LOKI_ENABLED")
	viper.BindEnv("monitoring.loki.url", "LOKI_URL")
}

// setDefaults sets default values for configuration
func setDefaults(serviceName string) {
	// Service defaults
	viper.SetDefault("service_name", serviceName)
	viper.SetDefault("environment", "development")
	viper.SetDefault("log_level", "info")

	// Port defaults based on service
	switch serviceName {
	case "gateway":
		viper.SetDefault("port", "8000")
	case "user":
		viper.SetDefault("port", "8001")
	case "product":
		viper.SetDefault("port", "8002")
	case "order":
		viper.SetDefault("port", "8003")
	default:
		viper.SetDefault("port", "8080")
	}

	// Database defaults
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	// Redis defaults
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// JWT defaults
	viper.SetDefault("jwt.expiration", "15m")
	viper.SetDefault("jwt.refresh_expiration", "168h") // 7 days

	// Service endpoint defaults
	viper.SetDefault("services.user_service.host", "localhost")
	viper.SetDefault("services.user_service.port", "8001")
	viper.SetDefault("services.user_service.url", "http://localhost:8001")
	
	viper.SetDefault("services.product_service.host", "localhost")
	viper.SetDefault("services.product_service.port", "8002")
	viper.SetDefault("services.product_service.url", "http://localhost:8002")
	
	viper.SetDefault("services.order_service.host", "localhost")
	viper.SetDefault("services.order_service.port", "8003")
	viper.SetDefault("services.order_service.url", "http://localhost:8003")
	
	viper.SetDefault("services.gateway_service.host", "localhost")
	viper.SetDefault("services.gateway_service.port", "8000")
	viper.SetDefault("services.gateway_service.url", "http://localhost:8000")

	// Monitoring defaults
	viper.SetDefault("monitoring.prometheus.enabled", true)
	viper.SetDefault("monitoring.prometheus.port", "9090")
	viper.SetDefault("monitoring.prometheus.path", "/metrics")
	
	viper.SetDefault("monitoring.grafana.enabled", true)
	viper.SetDefault("monitoring.grafana.url", "http://localhost:3000")
	viper.SetDefault("monitoring.grafana.user", "admin")
	viper.SetDefault("monitoring.grafana.pass", "admin")
	
	viper.SetDefault("monitoring.loki.enabled", true)
	viper.SetDefault("monitoring.loki.url", "http://localhost:3100")
}

// validateConfig validates that required configuration is present
func validateConfig(serviceName string) error {
	var missingVars []string

	// Check database URL (required for all services except gateway)
	if serviceName != "gateway" {
		if viper.GetString("database.url") == "" {
			missingVars = append(missingVars, "DATABASE_URL")
		}
	}

	// Check Redis address (required for all services)
	if viper.GetString("redis.addr") == "" {
		missingVars = append(missingVars, "REDIS_ADDR")
	}

	// Check JWT secret (required for user service and gateway)
	if (serviceName == "user" || serviceName == "gateway") && viper.GetString("jwt.secret") == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	// Check port
	if viper.GetString("port") == "" {
		missingVars = append(missingVars, "PORT")
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing required configuration for %s: %s", serviceName, strings.Join(missingVars, ", "))
	}

	return nil
}

// GetServiceConfig returns a service-specific configuration
func GetServiceConfig(serviceName string) (*ServiceConfig, error) {
	return LoadConfig(serviceName)
}

// IsDevelopment returns true if the environment is development
func (c *ServiceConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *ServiceConfig) IsProduction() bool {
	return c.Environment == "production"
}

// GetDatabaseURL returns the database URL
func (c *ServiceConfig) GetDatabaseURL() string {
	return c.Database.URL
}

// GetRedisAddr returns the Redis address
func (c *ServiceConfig) GetRedisAddr() string {
	return c.Redis.Addr
}

// GetJWTSecret returns the JWT secret
func (c *ServiceConfig) GetJWTSecret() string {
	return c.JWT.Secret
}

// GetUserServiceURL returns the user service URL
func (c *ServiceConfig) GetUserServiceURL() string {
	return c.Services.UserService.GetServiceURL()
}

// GetProductServiceURL returns the product service URL
func (c *ServiceConfig) GetProductServiceURL() string {
	return c.Services.ProductService.GetServiceURL()
}

// GetOrderServiceURL returns the order service URL
func (c *ServiceConfig) GetOrderServiceURL() string {
	return c.Services.OrderService.GetServiceURL()
}

// GetGatewayServiceURL returns the gateway service URL
func (c *ServiceConfig) GetGatewayServiceURL() string {
	return c.Services.GatewayService.GetServiceURL()
}

// getEnvironment returns the current environment
func getEnvironment() string {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}
	
	// Validate environment
	validEnvs := []string{"local", "development", "staging", "production"}
	for _, validEnv := range validEnvs {
		if env == validEnv {
			return env
		}
	}
	
	// Default to development if invalid
	return "development"
}