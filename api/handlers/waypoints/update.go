package waypoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type UpdateWaypointInput struct {
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	CustomerName  string  `json:"customer_name"`
	Notes         string  `json:"notes"`
	SequenceOrder int     `json:"sequence_order"`
}

func UpdateWaypoint(c *gin.Context) {
	waypointID := c.Param("id")

	// 1. Validar Input
	var input UpdateWaypointInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Buscar Waypoint
	var wp domains.Waypoint
	if err := database.DB.First(&wp, "id = ?", waypointID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Punto no encontrado"})
		return
	}

	// 3. Validar si la entrega ya se realizó
	if wp.IsCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se puede editar un punto ya visitado/completado"})
		return
	}

	// 4. Actualizar campos
	if input.Address != "" {
		wp.Address = input.Address
	}
	// Lat/Long pueden ser 0

	if input.Latitude != 0 && input.Longitude != 0 {
		wp.Latitude = input.Latitude
		wp.Longitude = input.Longitude
	}
	if input.CustomerName != "" {
		wp.CustomerName = input.CustomerName
	}
	if input.Notes != "" {
		wp.Notes = input.Notes
	}
	if input.SequenceOrder != 0 {
		wp.SequenceOrder = input.SequenceOrder
	}

	// 5. Guardar
	if err := database.DB.Save(&wp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar punto"})
		return
	}

	c.JSON(http.StatusOK, wp)
}
