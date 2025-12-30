package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tu-usuario/route-manager/api/database"
	"github.com/tu-usuario/route-manager/api/domains"
)

// UserIntentionInput captura la intención del usuario desde el formulario del Front
type UserIntentionInput struct {
	// CAMBIO 1: Quitamos 'required'. Permitimos que venga vacío (para Logins).
	// Mantenemos 'oneof' para que SI envían algo, sea válido.
	Role string `json:"role" binding:"oneof=admin driver"`
}

// RegisterUserFromGoogle sincroniza el usuario de Supabase con la BD local
func RegisterUserFromGoogle(c *gin.Context) {

	// 1. Recuperar datos seguros del contexto (Token)
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	emailVal, _ := c.Get("userEmail")
	nameVal, _ := c.Get("userName")
	avatarVal, _ := c.Get("userAvatar")
	verifiedVal, _ := c.Get("userVerified")

	// 2. Intentar leer la intención del usuario (Body JSON)
	var input UserIntentionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		// CAMBIO 2: Si el error es "EOF" significa que el body vino vacío.
		// Eso es normal en un Login, así que lo ignoramos por ahora.
		// Si es otro error (ej: role="hacker"), entonces sí fallamos.
		if err.Error() != "EOF" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Datos inválidos en la solicitud",
				"details": err.Error(),
			})
			return
		}
	}

	// 3. Conversiones seguras y Parsing de datos de Google
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

	// 4. Buscar usuario en Base de Datos
	var user domains.User
	result := database.DB.First(&user, "id = ?", uid)

	// CASO A: USUARIO NUEVO
	if result.RowsAffected == 0 {

		// Validación manual.
		if input.Role == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Para registrarte por primera vez debes seleccionar un rol (Soy Conductor o Soy Administrador).",
			})
			return
		}

		newUser := domains.User{
			ID:            uid,
			Email:         email,
			FullName:      fullName,
			AvatarURL:     avatarURL,
			EmailVerified: verified,

			Role: input.Role,

			Status: "inactive",

			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := database.DB.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando usuario: " + err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Solicitud de registro creada. Estado pendiente.",
			"user":    newUser,
		})

	} else {
		// CASO B: USUARIO EXISTENTE (Login recurrente)
		// Aquí NO miramos input.Role. Ignoramos si viene vacío o lleno.

		// Solo actualizamos datos cosméticos de Google (Nombre, Avatar)
		user.FullName = fullName
		user.AvatarURL = avatarURL
		user.EmailVerified = verified
		user.UpdatedAt = time.Now()

		if err := database.DB.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error actualizando usuario: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login exitoso",
			"user":    user,
		})
	}
}
