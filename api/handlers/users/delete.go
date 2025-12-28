package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Borrado lógico (Soft Delete)
	result := database.DB.Delete(&domains.User{}, "id = ?", id)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error eliminando usuario"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado correctamente"})
}
