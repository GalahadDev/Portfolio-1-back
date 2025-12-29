package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

func DeleteRoute(c *gin.Context) {
	routeID := c.Param("id")

	// 1. Buscar la ruta
	var route domains.Route
	if err := database.DB.First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// 2. REGLA DE NEGOCIO: Evitar borrar historial crítico

	if route.Status == "in_progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No puedes eliminar una ruta activa. Cancélala primero."})
		return
	}

	// 3. Borrado (Soft Delete gracias a gorm.DeletedAt en el modelo)

	tx := database.DB.Begin()

	if err := tx.Where("route_id = ?", route.ID).Delete(&domains.Waypoint{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando paradas"})
		return
	}

	if err := tx.Delete(&route).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando ruta"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "Ruta eliminada correctamente"})
}
