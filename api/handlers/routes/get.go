package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
	"github.com/tu-usuario/route-manager/api/services/storage"
)

func ListRoutes(c *gin.Context) {
	// Obtener datos del usuario actual
	userID, _ := c.Get("userID")

	// Consultar su rol en BD
	var user domains.User
	database.DB.Select("role").First(&user, "id = ?", userID)

	var routes []domains.Route
	query := database.DB.Preload("Waypoints").Preload("Driver")

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

// GetRouteByID obtiene el detalle de una ruta y firma las URLs de las fotos
func GetRouteByID(c *gin.Context) {
	routeID := c.Param("id")
	userID, _ := c.Get("userID")

	// 1. Buscar la ruta en BD
	var route domains.Route

	if err := database.DB.Preload("Waypoints").Preload("Driver").First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// 2. Seguridad: Verificar que el usuario tenga permiso para verla
	// Admin ve todo. Conductor solo ve la suya.
	var user domains.User
	database.DB.Select("role").First(&user, "id = ?", userID)

	if user.Role != "admin" {
		// Si es conductor, verificamos que sea SU ruta
		if route.DriverID == nil || route.DriverID.String() != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta ruta"})
			return
		}
	}

	// 3. Firmar URLs de las fotos (POD)
	storageSvc := storage.NewService()

	// Iteramos sobre los waypoints para buscar fotos
	for i := range route.Waypoints {
		// Obtenemos un puntero al waypoint actual para poder modificarlo
		wp := &route.Waypoints[i]

		// Si tiene una foto guardada (path interno)...
		if wp.ProofPhotoURL != nil && *wp.ProofPhotoURL != "" {
			// ...pedimos a Supabase una URL firmada válida por 1 hora
			signedURL, err := storageSvc.GetSignedURL(*wp.ProofPhotoURL)

			if err == nil {
				// Reemplazamos el path interno por la URL pública temporal
				// Esto solo afecta al JSON de respuesta, NO guarda en la BD
				wp.ProofPhotoURL = &signedURL
			} else {
				fmt.Printf("Error firmando foto waypoint %s: %v\n", wp.ID, err)
			}
		}
	}

	c.JSON(http.StatusOK, route)
}
