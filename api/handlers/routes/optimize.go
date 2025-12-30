package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"github.com/tu-usuario/route-manager/api/services/optimization"
)

func OptimizeRoute(c *gin.Context) {
	routeID := c.Param("id")

	// 1. Obtener la ruta y sus waypoints
	var route domains.Route
	// Preload es importante para traer los waypoints
	if err := database.DB.Preload("Waypoints").First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	if len(route.Waypoints) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Se necesitan al menos 3 puntos para optimizar"})
		return
	}

	// 2. Ejecutar el Algoritmo
	optimizedWaypoints := optimization.OptimizeRoute(route.Waypoints)

	// 3. Actualizar el orden (SequenceOrder) en la base de datos
	// GORM hace esto en una transacción para seguridad
	tx := database.DB.Begin()

	// Actualizamos distancia total estimada ya que estamos aquí
	newTotalDist := optimization.CalculateRouteDistance(optimizedWaypoints)

	// Actualizar cada waypoint con su nuevo orden
	for i, wp := range optimizedWaypoints {
		wp.SequenceOrder = i + 1 // Orden 1, 2, 3...
		if err := tx.Model(&domains.Waypoint{}).Where("id = ?", wp.ID).Update("sequence_order", wp.SequenceOrder).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando optimización"})
			return
		}
	}

	// Actualizar total km en la ruta
	if err := tx.Model(&route).Update("total_distance_km", newTotalDist).Error; err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message":           "Ruta optimizada exitosamente",
		"original_distance": route.TotalDistanceKm, // Distancia vieja (si existía)
		"new_distance":      newTotalDist,
		"optimized_order":   optimizedWaypoints,
	})
}
