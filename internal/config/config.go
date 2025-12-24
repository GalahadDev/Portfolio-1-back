package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	Environment string
	SupabaseURL string
}

// Load lee el archivo .env y carga la configuración
func Load() (*Config, error) {
	// Intentamos cargar .env, pero no damos error si falla
	_ = godotenv.Load()

	cfg := &Config{
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Environment: os.Getenv("ENV"),
		SupabaseURL: os.Getenv("SUPABASE_URL"),
	}

	if cfg.SupabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL es requerida")
	}

	// Validaciones básicas para no arrancar si falta algo crítico
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("la variable DATABASE_URL es obligatoria")
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	return cfg, nil
}
