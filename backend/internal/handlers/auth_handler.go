package handlers

import (
	"net/http"

	"github.com/beohoang98/moneyapp/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/auth/logout", h.handleLogout)
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	result, err := h.authService.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			respondError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondSingle(w, http.StatusOK, result)
}

func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
