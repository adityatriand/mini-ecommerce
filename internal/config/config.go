package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseUrl   string
	RedisAddr     string
	RedisPassword string
	Port          string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	return Config{
		DatabaseUrl:   os.Getenv("DATABASE_URL"),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		Port:          os.Getenv("PORT"),
	}
}
