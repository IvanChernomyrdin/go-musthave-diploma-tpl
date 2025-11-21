package tests

import (
	"errors"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	serviceTest "go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartService_GetUserByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1
	expectedUser := &models.User{
		ID:    userID,
		Login: "testuser",
	}

	mockRepo.EXPECT().
		GetUserByID(userID).
		Return(expectedUser, nil)

	user, err := service.GetUserByID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Login, user.Login)
}

func TestGofemartService_GetUserByID_InvalidID(t *testing.T) {
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
			user, err := service.GetUserByID(tt.userID)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Equal(t, handler.ErrInvalidUserID.Error(), err.Error())
		})
	}
}

func TestGofemartService_GetUserByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 999

	mockRepo.EXPECT().
		GetUserByID(userID).
		Return(nil, nil)

	user, err := service.GetUserByID(userID)

	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestGofemartService_GetUserByID_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	userID := 1

	mockRepo.EXPECT().
		GetUserByID(userID).
		Return(nil, errors.New("database error"))

	user, err := service.GetUserByID(userID)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "database error", err.Error())
}
