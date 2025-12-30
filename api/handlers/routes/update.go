package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type UpdateRouteInput struct {
	Name                 string     `json:"name"`
	ScheduledDate        *time.Time `json:"scheduled_date"`
	TotalDistanceKm      float64    `json:"total_distance_km"`
	EstimatedDurationMin int        `json:"estimated_duration_min"`
}

func UpdateRoute(c *gin.Context) {
	routeID := c.Param("id")

	// 1. Validar Body
	var input UpdateRouteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Buscar Ruta
	var route domains.Route
	if err := database.DB.First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// 3. REGLA DE NEGOCIO: No editar rutas que ya están en curso o terminadas
	if route.Status == "in_progress" || route.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se puede editar una ruta en curso o finalizada"})
		return
	}

	// 4. Actualizar campos (si vienen en el JSON)
	if input.Name != "" {
		route.Name = input.Name
	}
	if input.ScheduledDate != nil {
		route.ScheduledDate = input.ScheduledDate
	}
	if input.TotalDistanceKm != 0 {
		route.TotalDistanceKm = input.TotalDistanceKm
	}
	if input.EstimatedDurationMin != 0 {
		route.EstimatedDurationMin = input.EstimatedDurationMin
	}

	// 5. Guardar
	if err := database.DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar ruta"})
		return
	}

	c.JSON(http.StatusOK, route)
}
