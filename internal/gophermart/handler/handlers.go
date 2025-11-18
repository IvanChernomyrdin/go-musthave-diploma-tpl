package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	service "go-musthave-diploma-tpl/internal/gophermart/service"

	pgk "go-musthave-diploma-tpl/internal/pkg"
	logger "go-musthave-diploma-tpl/internal/runtime/logger"
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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	// нужно реализовать 200, 202, 400, 401, 409, 422, 500
	// пароверили метод text
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "неверный формат запроса", http.StatusBadRequest) // 400
		return
	}
	// провекрили cookie
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "пользователь не аутентифицирован", http.StatusUnauthorized) // 401
		return
	}
	// читаем тело
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// убираем пробелы из строки
	number := strings.TrimSpace(string(body))
	// строка должан быть не пустая
	if number == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}
	// проверяем что в строке все цифры
	if !pgk.ContainsOnlyDigits(number) {
		http.Error(w, "неверный формат номера", http.StatusUnprocessableEntity) //422
		return
	}
	// проходим алгоритмом луна
	if !pgk.ValidateLuhn(number) {
		http.Error(w, "неверный формат номера", http.StatusUnprocessableEntity)
		return
	}

	// Преобразуем string в int64
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusInternalServerError)
		return
	}

	// если всё удачно записываем в бд
	err = h.svc.CreateOrder(userIDint, number)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateOrder):
			w.WriteHeader(http.StatusOK) //200 если уже был загружен
			w.Write([]byte(err.Error()))
		case errors.Is(err, models.ErrOtherUserOrder): //409 если другой пользователь уже загрузил
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(err.Error()))
		default:
			http.Error(w, "внутренние ошибки сервера", http.StatusInternalServerError) //500
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("успешное создание заказа"))
}
