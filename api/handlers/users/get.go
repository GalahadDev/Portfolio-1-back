package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// GetMe devuelve los datos del usuario logueado actualmente
func GetMe(c *gin.Context) {
	// 1. Recuperar el ID que el middleware guardó
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autorizado"})
		return
	}

	// 2. Buscar en BD
	var user domains.User
	// Preload pre-carga las relaciones si las tuviéramos (ej: Rutas)
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ListUsers devuelve todos los usuarios (Útil para paneles de admin o selectores)
func ListUsers(c *gin.Context) {
	var users []domains.User

	// Filtros opcionales
	// Ejemplo: ?role=driver
	role := c.Query("role")
	query := database.DB

	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Buscar todos
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error obteniendo usuarios"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func GetUser(c *gin.Context) {
	id := c.Param("id") // Viene de la URL /users/:id

	var user domains.User
	// Buscamos por el ID recibido en la ruta
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}
