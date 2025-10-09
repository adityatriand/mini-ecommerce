package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseUrl       string
	RedisAddr         string
	RedisPassword     string
	Port              string
	TrustedProxies    []string
	JWTSecret         string
	JWTExpiration     time.Duration
	RefreshExpiration time.Duration
}

func Load() (Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvVariables()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return Config{}, fmt.Errorf("error reading config file: %w", err)
		}
	}

	setDefaults()

	var missingVars []string

	databaseUrl := viper.GetString("database.url")
	if databaseUrl == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}

	redisAddr := viper.GetString("redis.addr")
	if redisAddr == "" {
		missingVars = append(missingVars, "REDIS_ADDR")
	}

	port := viper.GetString("server.port")
	if port == "" {
		missingVars = append(missingVars, "PORT")
	}

	jwtSecret := viper.GetString("jwt.secret")
	if jwtSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	if len(missingVars) > 0 {
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missingVars, ", "))
	}

	trustedProxies := viper.GetStringSlice("server.trusted_proxies")
	if len(trustedProxies) == 0 {
		trustedProxies = []string{"127.0.0.1", "::1"}
	}

	jwtExpMinutes := viper.GetInt("jwt.exp_minutes")
	jwtExpiration := time.Duration(jwtExpMinutes) * time.Minute

	refreshExpHours := viper.GetInt("jwt.refresh_exp_hours")
	refreshExpiration := time.Duration(refreshExpHours) * time.Hour

	return Config{
		DatabaseUrl:       databaseUrl,
		RedisAddr:         redisAddr,
		RedisPassword:     viper.GetString("redis.password"),
		Port:              port,
		TrustedProxies:    trustedProxies,
		JWTSecret:         jwtSecret,
		JWTExpiration:     jwtExpiration,
		RefreshExpiration: refreshExpiration,
	}, nil
}

func bindEnvVariables() {
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("redis.addr", "REDIS_ADDR")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("server.trusted_proxies", "TRUSTED_PROXIES")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.exp_minutes", "JWT_EXP_MINUTES")
	viper.BindEnv("jwt.refresh_exp_hours", "REFRESH_EXP_HOURS")
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.trusted_proxies", []string{"127.0.0.1", "::1"})
	viper.SetDefault("jwt.exp_minutes", 15)
	viper.SetDefault("jwt.refresh_exp_hours", 168)
}
