package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	var missingVars []string

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		missingVars = append(missingVars, "DATABASE_URL")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		missingVars = append(missingVars, "REDIS_ADDR")
	}

	port := os.Getenv("PORT")
	if port == "" {
		missingVars = append(missingVars, "PORT")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		missingVars = append(missingVars, "JWT_SECRET")
	}

	if len(missingVars) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}

	proxies := os.Getenv("TRUSTED_PROXIES")
	var trustedProxies []string
	if proxies != "" {
		for _, proxy := range strings.Split(proxies, ",") {
			if trimmed := strings.TrimSpace(proxy); trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	} else {
		trustedProxies = []string{"127.0.0.1", "::1"}
	}

	jwtExpMinutes := getEnvAsInt("JWT_EXP_MINUTES", 15)
	jwtExpiration := time.Duration(jwtExpMinutes) * time.Minute

	refreshExpHours := getEnvAsInt("REFRESH_EXP_HOURS", 168)
	refreshExpiration := time.Duration(refreshExpHours) * time.Hour

	return Config{
		DatabaseUrl:       databaseUrl,
		RedisAddr:         redisAddr,
		RedisPassword:     os.Getenv("REDIS_PASSWORD"), // Optional, can be empty
		Port:              port,
		TrustedProxies:    trustedProxies,
		JWTSecret:         jwtSecret,
		JWTExpiration:     jwtExpiration,
		RefreshExpiration: refreshExpiration,
	}, nil
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Invalid value for %s: %s, using default: %d", key, valueStr, defaultValue)
		return defaultValue
	}

	return value
}
