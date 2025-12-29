package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	SupabaseURL string
	JWTSecret   string
}

func Load() (*Config, error) {
	// Carga archivo .env si estamos en local
	_ = godotenv.Load()

	// 1. Puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. URL de Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL es requerida")
	}

	// 3. Base de Datos
	var dbURL string

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	if dbUser != "" && dbPass != "" && dbHost != "" {

		if dbPort == "" {
			dbPort = "5432"
		}
		if dbName == "" {
			dbName = "postgres"
		}

		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require",
			dbUser, dbPass, dbHost, dbPort, dbName)

	} else {
		dbURL = os.Getenv("DATABASE_URL")
	}

	if dbURL == "" {
		return nil, fmt.Errorf("No se encontró configuración de BD (ni DB_* vars ni DATABASE_URL)")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET es requerido")
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		SupabaseURL: supabaseURL,
		JWTSecret:   jwtSecret,
	}, nil
}
