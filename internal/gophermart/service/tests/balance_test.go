package tests

import (
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	serviceMocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartService_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")

	tests := []struct {
		name           string
		userID         int
		setupMock      func()
		expectedResult models.Balance
		expectError    bool
	}{
		{
			name:   "Successful balance receipt",
			userID: 1,
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{
					Current:   500.5,
					Withdrawn: 42,
				}, nil)
			},
			expectedResult: models.Balance{
				Current:   500.5,
				Withdrawn: 42,
			},
			expectError: false,
		},
		{
			name:   "Wrong userID",
			userID: 0,
			setupMock: func() {
				// Не должно быть вызова к репозиторию
			},
			expectedResult: models.Balance{},
			expectError:    true,
		},
		{
			name:   "Repository error",
			userID: 2,
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(2).Return(models.Balance{}, assert.AnError)
			},
			expectedResult: models.Balance{},
			expectError:    true,
		},
		{
			name:   "Empty balance",
			userID: 3,
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(3).Return(models.Balance{
					Current:   0,
					Withdrawn: 0,
				}, nil)
			},
			expectedResult: models.Balance{
				Current:   0,
				Withdrawn: 0,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := svc.GetBalance(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
