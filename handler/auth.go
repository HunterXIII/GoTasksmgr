package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"tasksmgr/repo" // замените на ваш путь к repo
)

type AuthHandler struct {
	JWTSecret []byte
	Users     *repo.UserRepository
}

func NewAuthHandler(users *repo.UserRepository, secret string) *AuthHandler {
	return &AuthHandler{
		JWTSecret: []byte(secret),
		Users:     users,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			res, _ := json.Marshal("Invalid request")
			w.Write(res)
			return
		}

		user, err := h.Users.GetByUsername(req.Username)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("Invalid credentials")
			w.Write(res)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("Invalid credentials")
			w.Write(res)
			return
		}

		// Генерация JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.Id,
			"exp": time.Now().Add(3 * time.Minute).Unix(),
		})

		tokenString, err := token.SignedString(h.JWTSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal("Could not generate a token")
			w.Write(res)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		res, _ := json.Marshal(loginResponse{Token: tokenString})
		w.Write(res)
	}
}

func (h *AuthHandler) WhoAmI(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("UserID").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		res, _ := json.Marshal("No authorize")
		w.Write(res)
		return
	}
	res, _ := json.Marshal(fmt.Sprintf("Your UserID: %d\n", userID))
	w.Write(res)
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {

	parser := jwt.NewParser(
		jwt.WithLeeway(120*time.Second),         // допустимая погрешность времени
		jwt.WithValidMethods([]string{"HS256"}), // разрешённый алгоритм
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("No authorization")
			w.Write(res)
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("No authorization bearer")
			w.Write(res)
			return
		}

		tokenString := parts[1]
		token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return h.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("Invalid Token")
			w.Write(res)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("Invalid Token")
			w.Write(res)
			return
		}

		sub, ok := claims["sub"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			res, _ := json.Marshal("Invalid Token")
			w.Write(res)
			return
		}

		userID := int(sub.(float64)) // JWT числа приходят как float64

		ctx := context.WithValue(r.Context(), "UserID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value("UserID").(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		res, _ := json.Marshal("No authorize")
		w.Write(res)
		return
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(3 * time.Minute).Unix(),
	})

	tokenString, err := token.SignedString(h.JWTSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal("Could not generate a token")
		w.Write(res)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	res, _ := json.Marshal(loginResponse{Token: tokenString})
	w.Write(res)
}
