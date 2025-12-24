package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

// AuthMiddleware recibe la URL de Supabase para iniciar el JWKS
func AuthMiddleware(supabaseURL string) func(http.Handler) http.Handler {

	// 1. URL del conjunto de claves públicas (JWKS) de Supabase
	jwksURL := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL)

	// 2. Inicializar el JWKS
	// Esto se hace una sola vez para no llamar a Supabase en cada petición.
	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		log.Fatalf("❌ Error fatal: No se pudo inicializar JWKS desde Supabase: %v", err)
	}
	log.Println("✅ Sistema de Auth (JWKS) inicializado correctamente")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// A. Obtener Token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Autorización requerida", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// B. Parsear y Validar usando JWKS
			// keyfunc se encarga de buscar la clave pública correcta para este token
			token, err := jwt.Parse(tokenString, jwks.Keyfunc)

			if err != nil || !token.Valid {
				log.Printf("Token inválido: %v", err) // Log para debug
				http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
				return
			}

			// C. Extraer User ID
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Error en claims", http.StatusUnauthorized)
				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "Token sin User ID", http.StatusUnauthorized)
				return
			}

			// D. Continuar
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID
func GetUserID(ctx context.Context) (string, error) {
	val := ctx.Value(UserIDKey)
	if val == nil {
		return "", fmt.Errorf("usuario no encontrado")
	}
	return val.(string), nil
}
