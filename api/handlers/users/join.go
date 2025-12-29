package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type JoinFleetInput struct {
	Code string `json:"code" binding:"required"`
}

func JoinFleet(c *gin.Context) {
	userID, _ := c.Get("userID") // ID del conductor

	var input JoinFleetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "El código es requerido"})
		return
	}

	// 1. Buscar al JEFE dueño de ese código
	var manager domains.User
	// Buscamos que el código exista y que el dueño sea 'admin'
	if err := database.DB.First(&manager, "fleet_code = ? AND role = 'admin'", input.Code).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Código de flota inválido o inexistente"})
		return
	}

	// 2. Vincular al Conductor con el Jefe
	result := database.DB.Model(&domains.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"manager_id": manager.ID,
			"status":     "active",
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al unirse a la flota"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Te has unido exitosamente a la flota",
		"manager_name": manager.FullName,
		"status":       "active",
	})
}
