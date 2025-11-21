package tests

import (
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	serviceTest "go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartServiceWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	tests := []struct {
		name          string
		userID        int
		withdraw      models.WithdrawBalance
		setupMock     func()
		expectedError error
	}{
		{
			name:   "Successful withdrawal",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					Withdraw(1, models.WithdrawBalance{
						Order: "2377225624",
						Sum:   751,
					}).
					Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "Invalid order number",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					Withdraw(1, models.WithdrawBalance{
						Order: "2377225624",
						Sum:   751,
					}).
					Return(handler.ErrInvalidOrderNumber)
			},
			expectedError: handler.ErrInvalidOrderNumber,
		},
		{
			name:   "Insufficient funds",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					Withdraw(1, models.WithdrawBalance{
						Order: "2377225624",
						Sum:   751,
					}).
					Return(handler.ErrLackOfFunds)
			},
			expectedError: handler.ErrLackOfFunds,
		},
		{
			name:   "Database error",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					Withdraw(1, models.WithdrawBalance{
						Order: "2377225624",
						Sum:   751,
					}).
					Return(assert.AnError)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.setupMock()

			// Вызываем метод сервиса
			err := service.Withdraw(tt.userID, tt.withdraw)

			// Проверяем результат
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
