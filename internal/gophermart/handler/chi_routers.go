package httpserver

import (
	"net/http"

	middleware "go-musthave-diploma-tpl/internal/gophermart/middleware"
	service "go-musthave-diploma-tpl/internal/gophermart/service"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler, svc *service.GofemartService) http.Handler {
	r := chi.NewRouter()

	// логгер запросов
	r.Use(middleware.LoggerMiddleware())

	// публичные маршруты
	r.Post("/api/user/register", h.Register)
	r.Post("/api/user/login", h.Login)

	// защищённые маршруты
	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			// подключаем проверку cookie
			r.Use(middleware.CookieMiddleware(svc))
			// проверка cookie
			r.Get("/test-auth", h.TestAuth)

			r.Route("/orders", func(r chi.Router) {
				// загрузка пользователем номера заказа для расчёта
				r.Post("/", h.CreateOrder)
				// получение списка загруженных пользователем номеров заказов, статусов их обработки и информации о начислениях
				// r.Get("/", h.GetOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				// получение текущего баланса счёта баллов лояльности пользователя
				// r.Get("/", h.GetBalance)
				// запрос на списание баллов с накопительного счёта в счёт оплаты нового заказа
				// r.Post("/withdraw", h.BalanceWithdraw)
			})
			// получение информации о выводе средств с накопительного счёта пользователем
			// r.Get("/withdrawals", h.Withdrawals)
		})
	})
	return r
}
