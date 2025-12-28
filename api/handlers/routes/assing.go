package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type AssignDriverInput struct {
	DriverID string `json:"driver_id" binding:"required"`
}

func AssignDriver(c *gin.Context) {
	routeID := c.Param("id")
	var input AssignDriverInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar UUIDs
	driverUUID, err := uuid.Parse(input.DriverID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de conductor inválido"})
		return
	}

	// Verificar que la ruta existe
	var route domains.Route
	if err := database.DB.First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// Verificar que el conductor existe y es conductor
	var driver domains.User
	if err := database.DB.First(&driver, "id = ? AND role = 'driver'", driverUUID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El usuario no existe o no es un conductor"})
		return
	}

	// Actualizar ruta
	route.DriverID = &driverUUID
	route.Status = "pending" // Cambia estado a pendiente de inicio

	if err := database.DB.Save(&route).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al asignar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ruta asignada a " + driver.FullName, "route": route})
}
