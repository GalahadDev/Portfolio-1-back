package routes

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"         // Ajusta a tu path real
	"github.com/tu-usuario/route-manager/api/domains"          // Ajusta a tu path real
	"github.com/tu-usuario/route-manager/api/services/storage" // Ajusta a tu path real
)

// ListRoutes lista las rutas aplicando filtros de seguridad según el rol
func ListRoutes(c *gin.Context) {
	// 1. Obtener ID del usuario desde el contexto (puesto por el middleware de auth)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// 2. Obtener el rol del usuario actual
	var user domains.User
	if err := database.DB.Select("id, role").First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado"})
		return
	}

	var routes []domains.Route
	// Preparamos la query base con los preloads necesarios
	query := database.DB.Preload("Waypoints").Preload("Driver")

	// 3. LÓGICA DE NEGOCIO SEGÚN ROL
	switch user.Role {
	case "super_admin":
		// CASO 1: Super Admin
		// Ve TODO. No aplicamos filtros restrictivos de ID.
		// Solo aplicamos filtro de estado si viene en la URL
		status := c.Query("status")
		if status != "" {
			query = query.Where("status = ?", status)
		}

	case "admin":
		// CASO 2: Admin
		// Solo ve las rutas que ÉL creó (usando creator_id)
		query = query.Where("creator_id = ?", user.ID)

		// Filtro de estado opcional
		status := c.Query("status")
		if status != "" {
			query = query.Where("status = ?", status)
		}

	default:
		// CASO 3: Conductor (o cualquier otro rol)
		// Solo ve las rutas donde él es el conductor asignado
		query = query.Where("driver_id = ?", user.ID)
	}

	// 4. Ejecutar consulta
	if err := query.Find(&routes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error listando rutas"})
		return
	}

	c.JSON(http.StatusOK, routes)
}

// GetRouteByID obtiene el detalle de una ruta, verifica permisos y firma las URLs de las fotos
func GetRouteByID(c *gin.Context) {
	routeID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario no autenticado"})
		return
	}

	// 1. Buscar la ruta en BD
	var route domains.Route
	// Es importante traer creator_id y driver_id para validar permisos
	if err := database.DB.Preload("Waypoints").Preload("Driver").First(&route, "id = ?", routeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ruta no encontrada"})
		return
	}

	// 2. Obtener datos del usuario para verificar rol
	var user domains.User
	if err := database.DB.Select("id, role").First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuario inválido"})
		return
	}

	// 3. Seguridad: Verificar que el usuario tenga permiso para verla
	// Si es super_admin, saltamos esta validación (tiene acceso total)
	if user.Role != "super_admin" {

		if user.Role == "admin" {
			// Si es Admin, debe ser el CREADOR de la ruta
			// Comparamos los UUIDs (asumiendo que user.ID y route.CreatorID son compatibles o strings)
			if route.CreatorID.String() != user.ID.String() {
				c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta ruta (no eres el creador)"})
				return
			}
		} else {
			// Si es Conductor, debe ser el conductor ASIGNADO
			if route.DriverID == nil || route.DriverID.String() != user.ID.String() {
				c.JSON(http.StatusForbidden, gin.H{"error": "No tienes permiso para ver esta ruta"})
				return
			}
		}
	}

	// 4. Firmar URLs de las fotos (Lógica de Supabase Storage)
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
				// Logueamos el error pero no fallamos la petición completa
				fmt.Printf("Error firmando foto waypoint %s: %v\n", wp.ID, err)
			}
		}
	}

	c.JSON(http.StatusOK, route)
}
