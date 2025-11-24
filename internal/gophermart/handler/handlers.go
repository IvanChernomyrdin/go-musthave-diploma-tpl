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
	"go-musthave-diploma-tpl/internal/gophermart/service"

	pgk "go-musthave-diploma-tpl/pkg"
	logger "go-musthave-diploma-tpl/pkg/runtime/logger"
)

var castomLogger = logger.NewHTTPLogger().Logger.Sugar()

type Handler struct {
	svc *service.GofemartService
}

func NewHandler(svc *service.GofemartService) *Handler {
	return &Handler{svc: svc}
}

// ====================== REGISTER ======================
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, `{"error":"content-type must be application/json"}`, http.StatusBadRequest)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+ErrInvalidJSONFormat.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"`+ErrLoginAndPasswordRequired.Error()+`"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.RegisterUser(req.Login, req.Password)
	if err != nil {
		switch err.Error() {
		case "login already exists":
			http.Error(w, `{"error":"login already taken"}`, http.StatusConflict)
		case "login and password are required":
			http.Error(w, `{"error":"login and password are required"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	middleware.SetEncryptedCookie(w, strconv.Itoa(user.ID))
	w.WriteHeader(http.StatusOK)
}

// ====================== LOGIN ======================
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, `{"error":"content-type must be application/json"}`, http.StatusBadRequest)
		return
	}

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"`+ErrInvalidJSONFormat.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, `{"error":"`+ErrLoginAndPasswordRequired.Error()+`"}`, http.StatusBadRequest)
		return
	}

	user, err := h.svc.LoginUser(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidLoginOrPassword) {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		} else {
			castomLogger.Infof(ErrInternalServerError.Error(), err.Error())
			http.Error(w, `{"error":"`+ErrInternalServerError.Error()+`"}`, http.StatusInternalServerError)
		}
		return
	}

	middleware.SetEncryptedCookie(w, strconv.Itoa(user.ID))
	w.WriteHeader(http.StatusOK)
}

// ====================== CREATE ORDER ======================
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Получаем ID пользователя из контекста
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, `{"error":"user is not authenticated"}`, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"failed to read request body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, `{"error":"order number is required"}`, http.StatusBadRequest)
		return
	}

	if !pgk.ContainsOnlyDigits(orderNumber) || !pgk.ValidateLuhn(orderNumber) {
		http.Error(w, `{"error":"invalid order number"}`, http.StatusUnprocessableEntity)
		return
	}

	userIDint, _ := strconv.Atoi(userID)

	// Создаём заказ в базе
	err = h.svc.CreateOrder(userIDint, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, ErrDuplicateOrder):
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "order already uploaded"})
		case errors.Is(err, ErrOtherUserOrder):
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "order already uploaded by another user"})
		default:
			castomLogger.Infof(ErrInternalServerError.Error(), err.Error())
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "order accepted for processing"})
}

// ====================== GET ORDERS ======================
func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, ErrUserIsNotAuthenticated.Error(), http.StatusUnauthorized)
		return
	}

	userIDint, _ := strconv.Atoi(userID)
	result, err := h.svc.GetOrders(userIDint)
	if err != nil {
		http.Error(w, ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}

	if len(result) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// ====================== GET BALANCE ======================
func (h *Handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+ErrUserIsNotAuthenticated.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusInternalServerError)
		return
	}
	result, err := h.svc.GetBalance(userIDint)
	if err != nil {
		http.Error(w, `{"error":"`+ErrInternalServerError.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// ====================== WITHDRAW ======================
func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+ErrUserIsNotAuthenticated.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, `{"error":"content-type must be application/json"}`, http.StatusBadRequest)
		return
	}

	var withdraw models.WithdrawBalance
	if err := json.NewDecoder(r.Body).Decode(&withdraw); err != nil {
		http.Error(w, `{"error":"`+ErrInvalidJSONFormat.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if withdraw.Sum <= 0 {
		http.Error(w, `{"error":"sum must be positive"}`, http.StatusBadRequest)
		return
	}
	if !pgk.ContainsOnlyDigits(withdraw.Order) || !pgk.ValidateLuhn(withdraw.Order) {
		http.Error(w, `{"error":"invalid order number"}`, http.StatusUnprocessableEntity)
		return
	}

	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusInternalServerError)
		return
	}
	err = h.svc.Withdraw(userIDint, withdraw)
	if err != nil {
		switch err {
		case ErrInvalidOrderNumber:
			http.Error(w, `{"error":"`+ErrInvalidOrderNumber.Error()+`"}`, http.StatusUnprocessableEntity)
		case ErrLackOfFunds:
			http.Error(w, `{"error":"`+ErrLackOfFunds.Error()+`"}`, http.StatusPaymentRequired)
		default:
			castomLogger.Infof(ErrInternalServerError.Error(), err.Error())
			http.Error(w, `{"error":"`+ErrInternalServerError.Error()+`"}`, http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct{}{})
}

// ====================== WITHDRAWALS ======================
func (h *Handler) Withdrawals(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+ErrUserIsNotAuthenticated.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	userIDint, _ := strconv.Atoi(userID)
	withdrawals, err := h.svc.Withdrawals(userIDint)
	if err != nil {
		http.Error(w, `{"error":"`+ErrInternalServerError.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(withdrawals)
}
