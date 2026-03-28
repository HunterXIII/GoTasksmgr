package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"tasksmgr/contextx"
	"tasksmgr/repo"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	JWTSecrets [][]byte
	Users      *repo.UserRepository
}

func NewAuthHandler(users *repo.UserRepository) *AuthHandler {
	return &AuthHandler{
		JWTSecrets: [][]byte{
			[]byte("CURRENTSECRET"),
			[]byte("OLDSECRET"),
		},
		Users: users,
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

		tokenString, err := token.SignedString(h.JWTSecrets[0])
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
		var token *jwt.Token
		var err error

		for _, secret := range h.JWTSecrets {

			token, err = parser.Parse(tokenString, func(token *jwt.Token) (any, error) {
				return secret, nil
			})

			if err == nil && token.Valid {
				break
			}
		}

		if err != nil || token == nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
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

		ctx := context.WithValue(r.Context(), "UserID", userID) // через константу или отдельный тип

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

	tokenString, err := token.SignedString(h.JWTSecrets[0])
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

func (h *AuthHandler) JWTAuthInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (
	any,
	error,
) {
	md, _ := metadata.FromIncomingContext(ctx)
	authHeader := strings.Join(md.Get("authorization"), "")
	// fmt.Println(authHeader)
	if authHeader == "" {
		return nil, status.Error(codes.Unauthenticated, "No authorization")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, status.Error(codes.Unauthenticated, "Not bearer authorization")
	}

	parser := jwt.NewParser(
		jwt.WithLeeway(120*time.Second),
		jwt.WithValidMethods([]string{"HS256"}),
	)

	tokenString := parts[1]
	var token *jwt.Token
	var err error

	for _, secret := range h.JWTSecrets {

		token, err = parser.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return secret, nil
		})

		if err == nil && token.Valid {
			break
		}
	}

	if err != nil || token == nil || !token.Valid {
		return nil, status.Error(codes.Unauthenticated, "Token is invalid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Token is invalid")
	}

	sub, ok := claims["sub"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Token is invalid")
	}

	userID := int(sub.(float64))

	ctx = context.WithValue(ctx, contextx.UserIDKey{}, userID)
	return handler(ctx, req)
}
