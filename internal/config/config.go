package config

import (
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

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
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
		DatabaseUrl:       os.Getenv("DATABASE_URL"),
		RedisAddr:         os.Getenv("REDIS_ADDR"),
		RedisPassword:     os.Getenv("REDIS_PASSWORD"),
		Port:              os.Getenv("PORT"),
		TrustedProxies:    trustedProxies,
		JWTSecret:         os.Getenv("JWT_SECRET"),
		JWTExpiration:     jwtExpiration,
		RefreshExpiration: refreshExpiration,
	}
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
