package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"afs-todo-backend/internal/middleware"
	"afs-todo-backend/internal/models"
	"afs-todo-backend/internal/users"

	"golang.org/x/crypto/bcrypt"
)

const minPasswordLength = 8

type UserRepository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, name string, email string, passwordHash string) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.UserWithPassword, error)
	FindByID(ctx context.Context, id int64) (models.User, error)
}

type Handler struct {
	users  UserRepository
	tokens *TokenManager
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type authResponse struct {
	User userResponse `json:"user"`
}

type messageResponse struct {
	Message string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(userRepository UserRepository, tokenManager *TokenManager) *Handler {
	return &Handler{
		users:  userRepository,
		tokens: tokenManager,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	request.Name = strings.TrimSpace(request.Name)
	request.Email = normalizeEmail(request.Email)

	if message := validateRegisterRequest(request); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	exists, err := h.users.EmailExists(r.Context(), request.Email)
	if err != nil {
		log.Printf("check register email failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}
	if exists {
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("password hashing failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	user, err := h.users.Create(r.Context(), request.Name, request.Email, string(passwordHash))
	if err != nil {
		if errors.Is(err, users.ErrEmailAlreadyExists) {
			writeError(w, http.StatusConflict, "email already exists")
			return
		}
		log.Printf("register user failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{User: toUserResponse(user)})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	request.Email = normalizeEmail(request.Email)

	if message := validateLoginRequest(request); message != "" {
		writeError(w, http.StatusBadRequest, message)
		return
	}

	user, err := h.users.FindByEmail(r.Context(), request.Email)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		log.Printf("find login user failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, expiresAt, err := h.tokens.Generate(user.ID)
	if err != nil {
		log.Printf("jwt generation failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not log in")
		return
	}

	setAuthCookie(w, token, expiresAt)
	writeJSON(w, http.StatusOK, authResponse{User: toUserResponse(user.User)})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	writeJSON(w, http.StatusOK, messageResponse{Message: "logged out"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.users.FindByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, users.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		log.Printf("find current user failed: %v", err)
		writeError(w, http.StatusInternalServerError, "could not get current user")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{User: toUserResponse(user)})
}

func validateRegisterRequest(request registerRequest) string {
	if request.Name == "" {
		return "name is required"
	}
	if !isValidEmail(request.Email) {
		return "valid email is required"
	}
	if len(request.Password) < minPasswordLength {
		return "password must be at least 8 characters"
	}
	return ""
}

func validateLoginRequest(request loginRequest) string {
	if !isValidEmail(request.Email) {
		return "valid email is required"
	}
	if request.Password == "" {
		return "password is required"
	}
	return ""
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(destination)
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}

	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func setAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func toUserResponse(user models.User) userResponse {
	return userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: message})
}
