package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware valida el token de Supabase y extrae metadatos de Google
func AuthMiddleware(supabaseURL string) gin.HandlerFunc {

	// 1. Validaciones iniciales
	if supabaseURL == "" {
		log.Fatal("❌ Error fatal: Supabase URL vacía")
	}

	// 2. Configurar JWKS (Claves públicas de Supabase)
	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL)
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("❌ Error fatal inicializando JWKS: %v", err)
	}

	return func(c *gin.Context) {
		// A. Obtener el Token del Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Autorización requerida"})
			return
		}

		// Limpiar prefijo "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Formato de token inválido"})
			return
		}

		// B. Validar la firma del Token con JWKS
		token, err := jwt.Parse(tokenString, jwks.Keyfunc)

		if err != nil || !token.Valid {
			log.Printf("⚠️ Token rechazado: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido o expirado"})
			return
		}

		// C. Extraer Datos
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 1. ID de Usuario
			userID, okID := claims["sub"].(string)
			if !okID {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token sin User ID"})
				return
			}

			// 2. Email
			email, _ := claims["email"].(string)

			// 3. Metadatos (Nombre, Avatar, Verificado)
			var name, avatar string
			var verified bool

			if meta, ok := claims["user_metadata"].(map[string]interface{}); ok {
				if n, found := meta["full_name"].(string); found {
					name = n
				}
				if a, found := meta["avatar_url"].(string); found {
					avatar = a
				}
				if v, found := meta["email_verified"].(bool); found {
					verified = v
				}
			}

			// D. Guardar en el Contexto de Gin
			c.Set("userID", userID)
			c.Set("userEmail", email)
			c.Set("userName", name)
			c.Set("userAvatar", avatar)
			c.Set("userVerified", verified)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Error procesando claims"})
			return
		}
	}
}
