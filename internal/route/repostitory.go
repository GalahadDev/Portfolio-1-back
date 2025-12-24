package route

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tu-usuario/route-manager/internal/domain"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create guarda la cabecera de la ruta y todos sus waypoints en una sola transacción
func (r *Repository) Create(ctx context.Context, route *domain.Route) error {
	// 1. Iniciar Transacción
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	// Si algo falla al final, hacemos Rollback automático
	defer tx.Rollback(ctx)

	// 2. Insertar Cabecera
	queryRoute := `
		INSERT INTO public.routes (id, creator_id, name, status, scheduled_date, total_distance_km, estimated_duration_min, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	_, err = tx.Exec(ctx, queryRoute,
		route.ID, route.CreatorID, route.Name, route.Status, route.Date,
		route.TotalDistanceKM, route.EstimatedTimeMin,
	)
	if err != nil {
		return fmt.Errorf("error insertando ruta: %w", err)
	}

	// 3. Insertar Waypoints uno por uno
	queryWaypoint := `
		INSERT INTO public.waypoints (id, route_id, sequence_order, address, latitude, longitude, customer_name, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	for _, wp := range route.Waypoints {
		_, err := tx.Exec(ctx, queryWaypoint,
			wp.ID, route.ID, wp.SequenceOrder, wp.Address, wp.Latitude, wp.Longitude, wp.CustomerName, wp.Notes,
		)
		if err != nil {
			return fmt.Errorf("error insertando waypoint %d: %w", wp.SequenceOrder, err)
		}
	}

	// 4. Confirmar Transacción
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error haciendo commit de la ruta: %w", err)
	}

	return nil
}
