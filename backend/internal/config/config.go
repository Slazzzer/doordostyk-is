package config

import (
	"errors"
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	HTTPPort    string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		HTTPPort:    os.Getenv("HTTP_PORT"),
	}
	if cfg.DatabaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		log.Println("WARN: JWT_SECRET is empty, using insecure default")
		cfg.JWTSecret = "insecure-default-secret"
	}
	if cfg.HTTPPort == "" {
		cfg.HTTPPort = "8080"
	}
	return cfg, nil
}
