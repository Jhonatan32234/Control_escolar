package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"
const RolKey contextKey = "rol"

// AuthMiddleware verifica el token JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token de autorización requerido", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			http.Error(w, "Formato de token inválido. Use: Bearer <token>", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Token inválido o expirado", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Claims inválidos en el token", http.StatusUnauthorized)
			return
		}

		userID := claims["user_id"]
		rol := claims["rol"]

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, RolKey, rol)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware verifica que el usuario tenga el rol requerido
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rol, ok := r.Context().Value(RolKey).(string)
			if !ok {
				http.Error(w, "Rol no encontrado en el contexto", http.StatusForbidden)
				return
			}

			allowed := false
			for _, allowedRole := range allowedRoles {
				if rol == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "No tienes permisos para realizar esta acción", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
