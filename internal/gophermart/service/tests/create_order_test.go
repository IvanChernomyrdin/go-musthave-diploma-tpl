package tests

import (
	"errors"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	serviceTest "go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartService_CreateOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	orderNumber := "12345678903"

	mockRepo.EXPECT().
		CreateOrder(userID, orderNumber).
		Return(nil)

	err := service.CreateOrder(userID, orderNumber)

	assert.NoError(t, err)
}

func TestGofemartService_CreateOrder_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	tests := []struct {
		name   string
		userID int
	}{
		{"Zero ID", 0},
		{"Negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CreateOrder(tt.userID, "12345678903")

			assert.Error(t, err)
			assert.Equal(t, "Invalid user ID", err.Error())
		})
	}
}

func TestGofemartService_CreateOrder_EmptyOrderNumber(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1

	err := service.CreateOrder(userID, "")

	assert.Error(t, err)
	assert.Equal(t, "Order number is required", err.Error())
}

func TestGofemartService_CreateOrder_DuplicateOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	orderNumber := "12345678903"

	mockRepo.EXPECT().
		CreateOrder(userID, orderNumber).
		Return(handler.ErrDuplicateOrder)

	err := service.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, handler.ErrDuplicateOrder, err)
}

func TestGofemartService_CreateOrder_OtherUserOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	orderNumber := "12345678903"

	mockRepo.EXPECT().
		CreateOrder(userID, orderNumber).
		Return(handler.ErrOtherUserOrder)

	err := service.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, handler.ErrOtherUserOrder, err)
}

func TestGofemartService_CreateOrder_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	orderNumber := "12345678903"

	mockRepo.EXPECT().
		CreateOrder(userID, orderNumber).
		Return(errors.New("database error"))

	err := service.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
}
