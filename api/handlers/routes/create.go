package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// WaypointDTO: Lo que viene dentro del array de waypoints
type WaypointDTO struct {
	Address       string  `json:"address" binding:"required"`
	Latitude      float64 `json:"latitude" binding:"required"`
	Longitude     float64 `json:"longitude" binding:"required"`
	SequenceOrder int     `json:"sequence_order" binding:"required"`
	CustomerName  string  `json:"customer_name"`
	Notes         string  `json:"notes"`
}

// CreateRouteInput: El JSON completo que envía el Frontend
type CreateRouteInput struct {
	Name                 string        `json:"name" binding:"required"`
	ScheduledDate        *time.Time    `json:"scheduled_date"`
	TotalDistanceKm      int           `json:"total_distance_km"`
	EstimatedDurationMin int           `json:"estimated_duration_min"`
	Waypoints            []WaypointDTO `json:"waypoints" binding:"required,min=1"` // Mínimo 1 punto
}

func CreateRoute(c *gin.Context) {
	// 1. Obtener ID del Creador (Admin) del contexto
	creatorIDStr, _ := c.Get("userID")
	creatorUUID, _ := uuid.Parse(creatorIDStr.(string))

	// 2. Validar JSON
	var input CreateRouteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos: " + err.Error()})
		return
	}

	// 3. Mapear DTO a Entidades de Dominio

	var domainWaypoints []domains.Waypoint
	for _, wp := range input.Waypoints {
		domainWaypoints = append(domainWaypoints, domains.Waypoint{
			ID:            uuid.New(), // Generamos ID aquí
			Address:       wp.Address,
			Latitude:      wp.Latitude,
			Longitude:     wp.Longitude,
			SequenceOrder: wp.SequenceOrder,
			CustomerName:  wp.CustomerName,
			Notes:         wp.Notes,
			IsCompleted:   false,
		})
	}

	// Preparamos la Ruta
	newRoute := domains.Route{
		ID:                   uuid.New(),
		CreatorID:            creatorUUID, // El admin logueado
		DriverID:             nil,         // Se asigna después
		Name:                 input.Name,
		Status:               "draft",
		ScheduledDate:        input.ScheduledDate,
		TotalDistanceKm:      input.TotalDistanceKm,
		EstimatedDurationMin: input.EstimatedDurationMin,
		Waypoints:            domainWaypoints,
	}

	// 4. Guardar en Transacción
	if err := database.DB.Create(&newRoute).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo crear la ruta: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ruta creada con " + string(rune(len(domainWaypoints))) + " paradas",
		"route":   newRoute,
	})
}
