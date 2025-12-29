package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// RequireActiveUser verifica que el usuario tenga status="active" en la BD
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no identificado"})
			return
		}

		// Consultamos SOLO el estado
		var userStatus string

		// Hacemos una query ligera (SELECT status FROM users WHERE id = ?)
		result := database.DB.Model(&domains.User{}).
			Select("status").
			Where("id = ?", userID).
			Scan(&userStatus)

		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Usuario no encontrado en registros"})
			return
		}

		if userStatus != "active" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Tu cuenta está inactiva. Contacta al administrador para su aprobación.",
				"code":  "ACCOUNT_INACTIVE",
			})
			return
		}

		c.Next()
	}
}
