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

	pgk "go-musthave-diploma-tpl/pkg"
	logger "go-musthave-diploma-tpl/pkg/runtime/logger"
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
		http.Error(w, ErrInvalidJsonFormat.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, ErrLoginAndPasswordRequired.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.svc.RegisterUser(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case ErrLoginAndPasswordRequired.Error():
			http.Error(w, ErrLoginAndPasswordRequired.Error(), http.StatusBadRequest)
		case "login already exists":
			http.Error(w, "Login already taken", http.StatusConflict)
		default:
			castomLogger.Infof("Error registering user: %v\n", err)
			http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
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
		http.Error(w, ErrInvalidJsonFormat.Error(), http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, ErrLoginAndPasswordRequired.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.svc.LoginUser(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case "Invalid login or password":
			http.Error(w, err.Error(), http.StatusUnauthorized) // 401!
		default:
			castomLogger.Infof("Error logging in user: %v\n", err)
			http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
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
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Content-Type must be text/plain", http.StatusBadRequest)
		return
	}
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	// убираем пробелы из строки
	number := strings.TrimSpace(string(body))
	// строка должна быть не пустая
	if number == "" {
		http.Error(w, ErrOrderNumberRequired.Error(), http.StatusBadRequest)
		return
	}
	// проверяем что в строке все цифры
	if !pgk.ContainsOnlyDigits(number) {
		http.Error(w, ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		return
	}
	// прогоняем по алгоритму лунтика(luhn)
	if !pgk.ValidateLuhn(number) {
		http.Error(w, ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Преобразуем string в int64
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, ErrInvalidUserID.Error(), http.StatusInternalServerError)
		return
	}
	// если всё удачно записываем в бд
	err = h.svc.CreateOrder(userIDint, number)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateOrder): //200 если уже был загружен
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(err.Error()))
		case errors.Is(err, ErrOtherUserOrder): //409 если другой пользователь уже загрузил
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(err.Error()))
		default:
			http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError) //500
		}
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Successful order creation"))
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, ErrInvalidUserID.Error(), http.StatusInternalServerError)
		return
	}
	result, err := h.svc.GetOrders(userIDint)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(ErrInternalServerError.Error()))
	}
	if len(result) == 0 {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte("No information to answer"))
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, ErrInvalidUserID.Error(), http.StatusInternalServerError)
		return
	}

	result, err := h.svc.GetBalance(userIDint)
	if err != nil {
		http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, ErrInvalidUserID.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var withdraw models.WithdrawBalance
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		http.Error(w, ErrInvalidJsonFormat.Error(), http.StatusBadRequest)
		return
	}
	// прогоняем по луне и делаем проверки как в CreateOrder
	number := strings.TrimSpace(withdraw.Order)
	if number == "" {
		http.Error(w, ErrOrderNumberRequired.Error(), http.StatusBadRequest)
		return
	}
	if !pgk.ContainsOnlyDigits(number) {
		http.Error(w, ErrInvalidNumberFormat.Error(), http.StatusUnprocessableEntity)
		return
	}
	if !pgk.ValidateLuhn(number) {
		http.Error(w, ErrInvalidNumberFormat.Error(), http.StatusUnprocessableEntity)
		return
	}

	if withdraw.Sum <= 0 {
		http.Error(w, "Sum must be positive", http.StatusBadRequest)
		return
	}

	err = h.svc.Withdraw(userIDint, withdraw)
	if err != nil {
		switch err {
		case ErrInvalidOrderNumber:
			http.Error(w, ErrInvalidOrderNumber.Error(), http.StatusUnprocessableEntity)
		case ErrLackOfFunds:
			http.Error(w, ErrLackOfFunds.Error(), http.StatusPaymentRequired)
		default:
			http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Withdrawal successful"))
}

func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, ErrInvalidUserID.Error(), http.StatusInternalServerError)
		return
	}

	withdrawals, err := h.svc.Withdrawals(userIDint)
	if err != nil {
		http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(withdrawals)
}
