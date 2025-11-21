package tests

import (
	"errors"
	"testing"
	"time"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	serviceTest "go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartService_GetOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	expectedOrders := []models.Order{
		{
			UID:        1,
			UserID:     userID,
			Number:     "1234567890",
			Status:     "PROCESSED",
			Accrual:    100.5,
			UploadedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			UID:        2,
			UserID:     userID,
			Number:     "0987654321",
			Status:     "NEW",
			Accrual:    0,
			UploadedAt: time.Now(),
		},
	}

	mockRepo.EXPECT().
		GetOrders(userID).
		Return(expectedOrders, nil)

	orders, err := service.GetOrders(userID)

	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Len(t, orders, 2)
	assert.Equal(t, expectedOrders[0].Number, orders[0].Number)
	assert.Equal(t, expectedOrders[1].Status, orders[1].Status)
}

func TestGofemartService_GetOrders_EmptyOrders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	expectedOrders := []models.Order{}

	mockRepo.EXPECT().
		GetOrders(userID).
		Return(expectedOrders, nil)

	orders, err := service.GetOrders(userID)

	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Empty(t, orders)
}

func TestGofemartService_GetOrders_InvalidUserID(t *testing.T) {
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
			orders, err := service.GetOrders(tt.userID)

			assert.Error(t, err)
			assert.Nil(t, orders)
			assert.Equal(t, handler.ErrInvalidUserID.Error(), err.Error())
		})
	}
}

func TestGofemartService_GetOrders_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1

	mockRepo.EXPECT().
		GetOrders(userID).
		Return(nil, errors.New("database connection failed"))

	orders, err := service.GetOrders(userID)

	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.Equal(t, "database connection failed", err.Error())
}
