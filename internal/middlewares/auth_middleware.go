package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/ctxkeys"
)

func SuperOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(ctxkeys.UserRoleKey).(string)
		if role != "super" {
			http.Error(w, "Accès réservé au super administrateur", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authentification requise (Format: Bearer <token>)", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		secret := []byte(os.Getenv("JWT_SECRET"))

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {

			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("méthode de signature inattendue: %v", t.Header["alg"])
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Session invalide ou expirée", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Impossible de lire les données du token", http.StatusUnauthorized)
			return
		}

		uidRaw, okUID := claims["userID"]
		if !okUID || uidRaw == nil {
			http.Error(w, "Données utilisateur manquantes (userID)", http.StatusUnauthorized)
			return
		}

		userID, okTypeUID := uidRaw.(float64)
		if !okTypeUID {
			http.Error(w, "Format userID invalide", http.StatusUnauthorized)
			return
		}

		uuuidRaw, okUUID := claims["userUUID"]
		if !okUUID || uuuidRaw == nil {
			http.Error(w, "Données utilisateur manquantes (userUUID)", http.StatusUnauthorized)
			return
		}

		userUUID, okTypeUUID := uuuidRaw.(string)
		if !okTypeUUID {
			http.Error(w, "Format userUUID invalide (doit être une string)", http.StatusUnauthorized)
			return
		}

		oidRaw, okOID := claims["ownerID"]
		if !okOID || oidRaw == nil {
			http.Error(w, "Données propriétaire manquantes (ownerID)", http.StatusUnauthorized)
			return
		}
		ownerID, okTypeOID := oidRaw.(float64)
		if !okTypeOID {
			http.Error(w, "Format ownerID invalide", http.StatusUnauthorized)
			return
		}

		roleRaw, okRole := claims["role"]
		if !okRole || roleRaw == nil {
			http.Error(w, "Rôle utilisateur manquant", http.StatusUnauthorized)
			return
		}
		role, okTypeRole := roleRaw.(string)
		if !okTypeRole {
			http.Error(w, "Format rôle invalide", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, int(userID))
		ctx = context.WithValue(ctx, ctxkeys.UserUUIDKey, userUUID)
		ctx = context.WithValue(ctx, ctxkeys.OwnerIDKey, int(ownerID))
		ctx = context.WithValue(ctx, ctxkeys.UserRoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
