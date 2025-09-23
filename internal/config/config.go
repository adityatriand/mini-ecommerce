package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseUrl    string
	RedisAddr      string
	RedisPassword  string
	Port           string
	TrustedProxies []string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Parse trusted proxies from environment variable
	proxies := os.Getenv("TRUSTED_PROXIES")
	var trustedProxies []string
	if proxies != "" {
		// Split by comma and trim spaces
		for _, proxy := range strings.Split(proxies, ",") {
			if trimmed := strings.TrimSpace(proxy); trimmed != "" {
				trustedProxies = append(trustedProxies, trimmed)
			}
		}
	} else {
		// Default to localhost only if not specified
		trustedProxies = []string{"127.0.0.1", "::1"}
	}

	return Config{
		DatabaseUrl:    os.Getenv("DATABASE_URL"),
		RedisAddr:      os.Getenv("REDIS_ADDR"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		Port:           os.Getenv("PORT"),
		TrustedProxies: trustedProxies,
	}
}
