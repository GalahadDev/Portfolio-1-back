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

// FindAllByDriver busca todas las rutas asignadas a un conductor específico
func (r *Repository) FindAllByDriver(ctx context.Context, driverID string) ([]domain.Route, error) {
	// 1. Consultar las cabeceras de las rutas (Ordenadas por fecha más reciente)
	query := `
		SELECT id, creator_id, driver_id, name, status, scheduled_date, total_distance_km, estimated_duration_min, created_at
		FROM public.routes
		WHERE driver_id = $1
		ORDER BY scheduled_date DESC
	`
	rows, err := r.db.Query(ctx, query, driverID)
	if err != nil {
		return nil, fmt.Errorf("error buscando rutas del conductor: %w", err)
	}
	defer rows.Close()

	var routes []domain.Route

	// 2. Iterar sobre las rutas encontradas
	for rows.Next() {
		var rt domain.Route
		// Escaneamos los datos básicos
		if err := rows.Scan(
			&rt.ID, &rt.CreatorID, &rt.DriverID, &rt.Name, &rt.Status,
			&rt.Date, &rt.TotalDistanceKM, &rt.EstimatedTimeMin, &rt.CreatedAt,
		); err != nil {
			return nil, err
		}

		// 3. (Micro-optimización pendiente) Por ahora, haremos una consulta extra por cada ruta
		// para traer sus waypoints. En producción usaríamos un JOIN o array_agg, pero esto es más fácil de leer.
		wpQuery := `
			SELECT id, route_id, sequence_order, address, latitude, longitude, customer_name, notes, is_completed, completed_at
			FROM public.waypoints
			WHERE route_id = $1
			ORDER BY sequence_order ASC
		`
		wpRows, err := r.db.Query(ctx, wpQuery, rt.ID)
		if err != nil {
			return nil, fmt.Errorf("error buscando waypoints: %w", err)
		}

		var waypoints []domain.Waypoint
		for wpRows.Next() {
			var wp domain.Waypoint
			if err := wpRows.Scan(
				&wp.ID, &wp.RouteID, &wp.SequenceOrder, &wp.Address, &wp.Latitude, &wp.Longitude,
				&wp.CustomerName, &wp.Notes, &wp.IsCompleted, &wp.CompletedAt,
			); err != nil {
				wpRows.Close()
				return nil, err
			}
			waypoints = append(waypoints, wp)
		}
		wpRows.Close()

		// Asignamos los waypoints a la ruta y la agregamos a la lista final
		rt.Waypoints = waypoints
		routes = append(routes, rt)
	}

	return routes, nil
}

// UpdateStatus actualiza el estado de una ruta (ej: de assigned a in_progress)
func (r *Repository) UpdateStatus(ctx context.Context, routeID string, newStatus domain.RouteStatus) error {
	query := `
		UPDATE public.routes 
		SET status = $1 
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, newStatus, routeID)
	if err != nil {
		return fmt.Errorf("error actualizando estado de ruta: %w", err)
	}
	return nil
}
