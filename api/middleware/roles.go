package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// RequireRoles verifica que el usuario tenga uno de los roles permitidos
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Obtener el ID del usuario del contexto (puesto por AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
			return
		}

		// 2. Buscar al usuario en la BD para ver su ROL actual
		// (No confiamos solo en el token, consultamos la fuente de verdad: la DB)
		var user domains.User
		if err := database.DB.Select("role").First(&user, "id = ?", userID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en BD"})
			return
		}

		// 3. Verificar si el rol del usuario está en la lista permitida
		isAllowed := false
		for _, role := range allowedRoles {
			if user.Role == role {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Acceso denegado: No tienes permisos suficientes para esta acción",
			})
			return
		}

		// 4. Si tiene permiso
		c.Next()
	}
}
