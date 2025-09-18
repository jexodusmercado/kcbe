package middleware

import (
	"context"
	"fmt"
	"kabancount/internal/config"
	"kabancount/internal/store"
	"kabancount/internal/tokens"
	"kabancount/internal/utils"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type UserMiddleware struct {
	UserStore store.UserStore
}

type contextkey string

const UserContextKey = contextkey("user")

func SetUser(r *http.Request, u *store.User) *http.Request {
	ctx := context.WithValue(r.Context(), UserContextKey, u)
	return r.WithContext(ctx)
}

func GetUser(r *http.Request) *store.User {
	user, ok := r.Context().Value(UserContextKey).(*store.User)
	if !ok {
		return store.AnonymousUser
	}
	return user
}

func (um *UserMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = SetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Printf("Invalid authorization header format")
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authorization header format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {
			log.Printf("Empty token after Bearer prefix")
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authorization header format"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil {
			log.Printf("Error parsing token: %v", err)
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authentication token"})
			return
		}

		if !token.Valid {
			log.Printf("Invalid token: %v", err)
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authentication token"})
			return
		}

		user, err := um.UserStore.GetUserToken(tokens.ScopeAuth, tokenString)
		if err != nil {
			log.Printf("Error fetching user for token: %v", err)
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authentication token"})
			return
		}

		if user == nil {
			log.Printf("No user found for token")
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid authentication token"})
			return
		}

		r = SetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

func (um *UserMiddleware) RequireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "you must be authenticated to access this resource"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (um *UserMiddleware) RequireAdminUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUser(r)
		if user.IsAnonymous() {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "you must be authenticated to access this resource"})
			return
		}
		if user.Role != "admin" {
			utils.WriteJSON(w, http.StatusForbidden, utils.Envelope{"error": "you do not have permission to access this resource"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
