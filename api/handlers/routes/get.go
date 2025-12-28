package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

func ListRoutes(c *gin.Context) {
	// Obtener datos del usuario actual
	userID, _ := c.Get("userID")

	// Consultar su rol en BD (o confiar en el token si lo añadimos al claim,
	// pero consultar BD es más seguro para cambios recientes)
	var user domains.User
	database.DB.Select("role").First(&user, "id = ?", userID)

	var routes []domains.Route
	query := database.DB.Preload("Waypoints").Preload("Driver") // Cargamos datos relacionados

	// LÓGICA DE NEGOCIO SEGÚN ROL
	if user.Role == "admin" {
		// Admin ve todo, opcionalmente filtrar por estado
		status := c.Query("status")
		if status != "" {
			query = query.Where("status = ?", status)
		}
	} else {
		// Conductor SOLO ve sus rutas asignadas
		query = query.Where("driver_id = ?", userID)
	}

	// Ejecutar consulta
	if err := query.Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando rutas"})
		return
	}

	c.JSON(http.StatusOK, routes)
}
