package domain

import (
	"time"
)

type RouteStatus string

const (
	RouteStatusDraft      RouteStatus = "draft"
	RouteStatusAssigned   RouteStatus = "assigned"
	RouteStatusInProgress RouteStatus = "in_progress"
	RouteStatusCompleted  RouteStatus = "completed"
	RouteStatusCancelled  RouteStatus = "cancelled"
)

// Route representa el viaje completo de un conductor
type Route struct {
	ID        string      `json:"id"`
	CreatorID string      `json:"creator_id"`
	DriverID  *string     `json:"driver_id,omitempty"` // Puede ser null al crearla
	Name      string      `json:"name"`
	Status    RouteStatus `json:"status"`
	Date      time.Time   `json:"scheduled_date"`

	// Datos calculados (para KPIs)
	TotalDistanceKM  int `json:"total_distance_km"`
	EstimatedTimeMin int `json:"estimated_time_min"`

	// Relación: Una ruta tiene muchos puntos de parada
	Waypoints []Waypoint `json:"waypoints,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// Waypoint es una parada específica dentro de la ruta
type Waypoint struct {
	ID            string `json:"id"`
	RouteID       string `json:"route_id"`
	SequenceOrder int    `json:"sequence_order"` // 1, 2, 3...

	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	CustomerName string `json:"customer_name"`
	Notes        string `json:"notes,omitempty"`

	IsCompleted bool       `json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
