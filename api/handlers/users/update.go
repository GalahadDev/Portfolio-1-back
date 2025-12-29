package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"github.com/tu-usuario/route-manager/api/utils"
)

type UpdateUserInput struct {
	FullName string `json:"full_name"`
	Role     string `json:"role"`   // driver, admin
	Status   string `json:"status"` // active, inactive
}

func UpdateUser(c *gin.Context) {
	targetID := c.Param("id")

	// 1. Obtener quien hace la petición (Super Admin)
	// (Asumimos que el middleware ya validó que quien llama es Super Admin o Admin)

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

		// Si lo estamos promoviendo a ADMIN y no tiene código, se lo generamos.
		if input.Role == "admin" && user.FleetCode == "" {
			user.FleetCode = utils.GenerateFleetCode()
		}
	}

	// 3. Guardar
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al actualizar usuario"})
		return
	}

	c.JSON(http.StatusOK, user)
}
