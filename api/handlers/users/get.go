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
	requestingUserID, _ := c.Get("userID")

	// 1. Saber quién está pidiendo la lista
	var currentUser domains.User
	database.DB.First(&currentUser, "id = ?", requestingUserID)

	var users []domains.User
	query := database.DB.Model(&domains.User{})

	// 2. FILTRO DE JERARQUÍA
	switch currentUser.Role {
	case "super_admin":
		// Super Admin ve a TODOS (Admins y Drivers)
	case "admin":
		// Admin normal SOLO ve a SUS conductores
		query = query.Where("manager_id = ?", currentUser.ID)
	default:
		// Un driver no debería poder listar usuarios
		c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para listar usuarios"})
		return
	}

	// Ejecutar
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error al obtener usuarios"})
		return
	}

	c.JSON(http.StatusOK, users)
}
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user domains.User
	// Buscamos por el ID recibido en la ruta
	if err := database.DB.First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}
