package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"github.com/tu-usuario/route-manager/api/utils"
)

// Input para actualizar
type UpdateUserInput struct {
	Role   string `json:"role"`
	Status string `json:"status"`
}

func UpdateUser(c *gin.Context) {
	targetID := c.Param("id")

	var input UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user domains.User
	if err := database.DB.First(&user, "id = ?", targetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// 2. Actualizar campos
	if input.Status != "" {
		user.Status = input.Status
	}

	if input.Role != "" {
		user.Role = input.Role

		// 1. Verificamos si es nil
		if input.Role == "admin" && user.FleetCode == nil {

			// 2. Generamos el código en una variable temporal
			newCode := utils.GenerateFleetCode()

			// 3. Asignamos la dirección de memoria (&) al puntero
			user.FleetCode = &newCode
		}
	}

	// 3. Guardar
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar usuario"})
		return
	}

	c.JSON(http.StatusOK, user)
}
