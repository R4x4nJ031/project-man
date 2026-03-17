package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	JWTIssuer string
	JWTAud    string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("JWT_SECRET not set")
	}

	issuer := os.Getenv("JWT_ISSUER")
	if issuer == "" {
		issuer = "project-man"
	}

	aud := os.Getenv("JWT_AUDIENCE")
	if aud == "" {
		aud = "project-man-users"
	}

	return &Config{
		Port:      port,
		JWTSecret: secret,
		JWTIssuer: issuer,
		JWTAud:    aud,
	}
}