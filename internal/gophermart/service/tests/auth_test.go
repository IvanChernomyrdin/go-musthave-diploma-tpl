package tests

import (
	"errors"
	"testing"

	"go-musthave-diploma-tpl/internal/gophermart/models"
	serviceTest "go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGofemartService_RegisterUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	// подготовка
	login, password := "newuser", "password123"
	expectedUser := &models.User{
		ID:    1,
		Login: login,
	}

	// что мы ожидаем
	mockRepo.EXPECT().
		CreateUser(login, password).
		Return(expectedUser, nil)

	// выполянем регистрацию
	user, err := service.RegisterUser(login, password)

	// сравниваем
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Login, user.Login)
}

func TestGofemartService_RegisterUser_EmptyCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"Empty login", "", "password123"},
		{"Empty password", "user", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.RegisterUser(tt.login, tt.password)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Equal(t, "login and password are required", err.Error())
		})
	}
}

func TestGofemartService_RegisterUser_LoginExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	login, password := "existinguser", "password123"

	mockRepo.EXPECT().
		CreateUser(login, password).
		Return(nil, errors.New("login already exists"))

	user, err := service.RegisterUser(login, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "login already exists", err.Error())
}

func TestGofemartService_LoginUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	login, password := "testuser", "correctpassword"
	expectedUser := &models.User{
		ID:    1,
		Login: login,
	}

	mockRepo.EXPECT().
		GetUserByLoginAndPassword(login, password).
		Return(expectedUser, nil)

	user, err := service.LoginUser(login, password)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, expectedUser.Login, user.Login)
}

func TestGofemartService_LoginUser_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	login, password := "testuser", "wrongpassword"

	mockRepo.EXPECT().
		GetUserByLoginAndPassword(login, password).
		Return(nil, nil)

	user, err := service.LoginUser(login, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "invalid login or password", err.Error())
}

func TestGofemartService_LoginUser_EmptyCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	tests := []struct {
		name     string
		login    string
		password string
	}{
		{"Empty login", "", "password123"},
		{"Empty password", "user", ""},
		{"Both empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.LoginUser(tt.login, tt.password)

			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Equal(t, "login and password are required", err.Error())
		})
	}
}

func TestGofemartService_LoginUser_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	service := serviceTest.NewGofemartService(mockRepo)

	login, password := "testuser", "password123"

	mockRepo.EXPECT().
		GetUserByLoginAndPassword(login, password).
		Return(nil, errors.New("database connection failed"))

	user, err := service.LoginUser(login, password)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "database connection failed", err.Error())
}
