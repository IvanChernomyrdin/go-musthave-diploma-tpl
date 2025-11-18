package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	serviceMocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRouter_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	router := handler.NewRouter(h, svc)

	t.Run("Полный цикл запроса", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/register", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		require.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})
}
