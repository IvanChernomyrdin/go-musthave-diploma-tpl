package tests

import (
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGofemartService_Withdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)

	// Фиксированное время для тестов
	now := time.Now()
	time1 := now.Add(-24 * time.Hour)
	time2 := now.Add(-12 * time.Hour)

	tests := []struct {
		name           string
		userID         int
		mockSetup      func()
		expectedResult []models.WithdrawBalance
		expectedError  error
	}{
		{
			name:   "Successful withdrawals retrieval",
			userID: 1,
			mockSetup: func() {
				expectedWithdrawals := []models.WithdrawBalance{
					{
						Order:       "2377225624",
						Sum:         751.50,
						ProcessedAt: time1,
					},
					{
						Order:       "49927398716",
						Sum:         500.25,
						ProcessedAt: time2,
					},
				}
				mockRepo.EXPECT().Withdrawals(1).Return(expectedWithdrawals, nil)
			},
			expectedResult: []models.WithdrawBalance{
				{
					Order:       "2377225624",
					Sum:         751.50,
					ProcessedAt: time1,
				},
				{
					Order:       "49927398716",
					Sum:         500.25,
					ProcessedAt: time2,
				},
			},
			expectedError: nil,
		},
		{
			name:   "No withdrawals for user",
			userID: 2,
			mockSetup: func() {
				mockRepo.EXPECT().Withdrawals(2).Return([]models.WithdrawBalance{}, nil)
			},
			expectedResult: []models.WithdrawBalance{},
			expectedError:  nil,
		},
		{
			name:   "Database error",
			userID: 3,
			mockSetup: func() {
				mockRepo.EXPECT().Withdrawals(3).Return(nil, assert.AnError)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
		{
			name:   "Nil withdrawals from repository",
			userID: 4,
			mockSetup: func() {
				mockRepo.EXPECT().Withdrawals(4).Return(nil, nil)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:   "Single withdrawal",
			userID: 5,
			mockSetup: func() {
				expectedWithdrawals := []models.WithdrawBalance{
					{
						Order:       "1234567890",
						Sum:         300.75,
						ProcessedAt: time1,
					},
				}
				mockRepo.EXPECT().Withdrawals(5).Return(expectedWithdrawals, nil)
			},
			expectedResult: []models.WithdrawBalance{
				{
					Order:       "1234567890",
					Sum:         300.75,
					ProcessedAt: time1,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.mockSetup()

			// Вызываем тестируемый метод
			result, err := svc.Withdrawals(tt.userID)

			// Проверяем ошибку
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Проверяем результат
			if tt.expectedResult != nil {
				require.NotNil(t, result)
				assert.Len(t, result, len(tt.expectedResult))

				for i, expected := range tt.expectedResult {
					assert.Equal(t, expected.Order, result[i].Order)
					assert.Equal(t, expected.Sum, result[i].Sum)
					assert.WithinDuration(t, expected.ProcessedAt, result[i].ProcessedAt, time.Second)
				}
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestGofemartService_Withdrawals_RepositoryCalledCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)

	t.Run("Repository method called with correct parameters", func(t *testing.T) {
		userID := 123
		expectedWithdrawals := []models.WithdrawBalance{
			{
				Order:       "1234567890",
				Sum:         100.0,
				ProcessedAt: time.Now(),
			},
		}

		// Проверяем что метод репозитория вызывается с правильным userID
		mockRepo.EXPECT().Withdrawals(userID).Return(expectedWithdrawals, nil)

		result, err := svc.Withdrawals(userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedWithdrawals, result)
	})

	t.Run("Error propagation from repository", func(t *testing.T) {
		userID := 456
		expectedError := assert.AnError

		mockRepo.EXPECT().Withdrawals(userID).Return(nil, expectedError)

		result, err := svc.Withdrawals(userID)

		assert.ErrorIs(t, err, expectedError)
		assert.Nil(t, result)
	})
}
