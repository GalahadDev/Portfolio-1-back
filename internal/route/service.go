package route

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/internal/domain"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreateRouteInput define qué necesitamos que nos envíe el Frontend
type CreateRouteInput struct {
	Name             string            `json:"name"`
	ScheduledDate    string            `json:"scheduled_date"` // YYYY-MM-DD
	Waypoints        []domain.Waypoint `json:"waypoints"`
	EstimatedKM      int               `json:"estimated_km"`
	EstimatedMinutes int               `json:"estimated_minutes"`
}

func (s *Service) CreateRoute(ctx context.Context, creatorID string, input CreateRouteInput) (*domain.Route, error) {
	// 1. Validaciones básicas
	if len(input.Waypoints) == 0 {
		return nil, errors.New("la ruta debe tener al menos una parada")
	}

	// Parsear fecha
	date, err := time.Parse("2006-01-02", input.ScheduledDate)
	if err != nil {
		return nil, errors.New("formato de fecha inválido (use YYYY-MM-DD)")
	}

	// 2. Construir el objeto de Dominio
	routeID := uuid.NewString()

	newRoute := &domain.Route{
		ID:               routeID,
		CreatorID:        creatorID,
		Name:             input.Name,
		Status:           domain.RouteStatusDraft,
		Date:             date,
		TotalDistanceKM:  input.EstimatedKM,
		EstimatedTimeMin: input.EstimatedMinutes,
		Waypoints:        make([]domain.Waypoint, len(input.Waypoints)),
	}

	// 3. Preparar los Waypoints
	for i, wp := range input.Waypoints {
		wp.ID = uuid.NewString()
		wp.RouteID = routeID
		wp.SequenceOrder = i + 1 // Orden 1, 2, 3...
		wp.IsCompleted = false
		newRoute.Waypoints[i] = wp
	}

	// 4. Guardar en Base de Datos
	if err := s.repo.Create(ctx, newRoute); err != nil {
		return nil, err
	}

	return newRoute, nil
}
