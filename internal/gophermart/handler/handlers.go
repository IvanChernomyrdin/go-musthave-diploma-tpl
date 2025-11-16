package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	service "go-musthave-diploma-tpl/internal/gophermart/service"

	logger "go-musthave-diploma-tpl/internal/gophermart/runtime/logger"
)

// логгер
var castomLogger = logger.NewHTTPLogger().Logger.Sugar()

type Handler struct {
	svc *service.GofemartService
}

func NewHandler(svc *service.GofemartService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.svc.RegisterUser(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case "login and password are required":
			http.Error(w, "Login and password are required", http.StatusBadRequest)
		case "login already exists":
			http.Error(w, "Login already taken", http.StatusConflict)
		default:
			castomLogger.Infof("Error registering user: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Используем публичную функцию из middleware cookie
	userIDStr := strconv.Itoa(user.ID)
	middleware.SetEncryptedCookie(w, userIDStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User successfully registered",
		"user_id": user.ID,
		"login":   user.Login,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var req models.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.svc.LoginUser(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case "login and password are required":
			http.Error(w, "Login and password are required", http.StatusBadRequest)
		case "invalid login or password":
			http.Error(w, "Invalid login or password", http.StatusUnauthorized) // 401!
		default:
			castomLogger.Infof("Error logging in user: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Устанавливаем аутентификационную куку
	userIDStr := strconv.Itoa(user.ID)
	middleware.SetEncryptedCookie(w, userIDStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Successfully logged in",
		"user_id": user.ID,
		"login":   user.Login,
	})
}

func (h *Handler) TestAuth(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (это делает CookieMiddleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Successfully authenticated with cookie",
		"user_id": userID,
	})
}
