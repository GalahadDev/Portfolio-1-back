package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

type UpdateUserInput struct {
	FullName string `json:"full_name"`
	Role     string `json:"role"`   // driver, admin
	Status   string `json:"status"` // active, inactive
}

func UpdateUser(c *gin.Context) {
	id := c.Param("id") // ID que viene en la URL: /users/:id
	var input UpdateUserInput

	// 1. Validar Body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Buscar Usuario
	var user domains.User
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	// 3. Actualizar campos (Solo si traen datos)
	if input.FullName != "" {
		user.FullName = input.FullName
	}
	if input.Role != "" {
		user.Role = input.Role
	}
	if input.Status != "" {
		user.Status = input.Status
	}

	// 4. Guardar
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo actualizar"})
		return
	}

	c.JSON(http.StatusOK, user)
}
