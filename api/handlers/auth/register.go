package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// RegisterUserFromGoogle sincroniza el usuario de Supabase con la BD local
func RegisterUserFromGoogle(c *gin.Context) {

	// 1. Recuperar datos del contexto
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	emailVal, _ := c.Get("userEmail")
	nameVal, _ := c.Get("userName")
	avatarVal, _ := c.Get("userAvatar")
	verifiedVal, _ := c.Get("userVerified")

	// 2. Conversiones seguras
	uid, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuario inválido"})
		return
	}

	email := ""
	if emailVal != nil {
		email = emailVal.(string)
	}
	fullName := ""
	if nameVal != nil {
		fullName = nameVal.(string)
	}
	avatarURL := ""
	if avatarVal != nil {
		avatarURL = avatarVal.(string)
	}
	verified := false
	if verifiedVal != nil {
		verified = verifiedVal.(bool)
	}

	// 3. Buscar usuario en Base de Datos
	var user domains.User
	result := database.DB.First(&user, "id = ?", uid)

	if result.RowsAffected == 0 {
		newUser := domains.User{
			ID:            uid,
			Email:         email,
			FullName:      fullName,
			AvatarURL:     avatarURL,
			EmailVerified: verified, // Guardamos si el email está verificado
			Role:          "driver", // Rol por defecto
			Status:        "inactive",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := database.DB.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando usuario: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Usuario registrado exitosamente",
			"user":    newUser,
		})

	} else {
		user.FullName = fullName
		user.AvatarURL = avatarURL
		user.EmailVerified = verified
		user.UpdatedAt = time.Now()

		// Guardamos los cambios
		if err := database.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Datos de usuario sincronizados",
			"user":    user,
		})
	}
}
