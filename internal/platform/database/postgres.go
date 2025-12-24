package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New inicializa el pool de conexiones a PostgreSQL
func New(databaseURL string) (*pgxpool.Pool, error) {
	// Configuramos el pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error parseando url de base de datos: %w", err)
	}

	config.MaxConns = 15               // Máximo conexiones simultáneas
	config.MinConns = 2                // Mínimo conexiones siempre vivas
	config.MaxConnLifetime = time.Hour // Reciclar conexiones cada hora

	// Contexto con timeout para no quedarnos colgados si la red falla
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Creamos el pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %w", err)
	}

	// Hacemos un Ping para verificar que realmente estamos conectados
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error haciendo ping a la base de datos: %w", err)
	}

	log.Println("✅ Conectado exitosamente a Supabase (PostgreSQL)")
	return pool, nil
}
