package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"` // in_progress, completed, cancelled
}

func UpdateRouteStatus(c *gin.Context) {
	routeID := c.Param("id")
	userID, _ := c.Get("userID") // ID del conductor logueado

	var input UpdateStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Buscar la ruta
	var route domains.Route
	if err := database.DB.First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// 2. SEGURIDAD: Verificar que la ruta pertenece al conductor
	if route.DriverID == nil || route.DriverID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para modificar esta ruta"})
		return
	}

	// 3. Validar estado
	validStatuses := map[string]bool{"pending": true, "in_progress": true, "completed": true, "cancelled": true}
	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Estado inválido"})
		return
	}

	// 4. Actualizar
	route.Status = input.Status

	if err := database.DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando estado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Estado actualizado", "status": route.Status})
}
