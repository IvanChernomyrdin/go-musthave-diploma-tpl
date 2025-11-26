package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	mock_handler "go-musthave-diploma-tpl/internal/accrual/handler/mocks"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_CreateProductReward(t *testing.T) {
	type args struct {
		match      string
		reward     float64
		rewardType string
	}
	type mockBehavior func(r *mock_handler.MockService, args args)

	tests := []struct {
		name                 string
		inputBody            string
		inputArgs            args
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "Ok",
			inputBody: `{"match":"12345","reward":10.5,"reward_type":"percent"}`,
			inputArgs: args{
				match:      "12345",
				reward:     10.5,
				rewardType: "percent",
			},
			mockBehavior: func(r *mock_handler.MockService, args args) {
				r.EXPECT().CreateProductReward(args.match, args.reward, args.rewardType).Return(nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "",
		},
		{
			name:                 "Wrong input",
			inputBody:            `{"match":"12345"`,
			inputArgs:            args{},
			mockBehavior:         func(r *mock_handler.MockService, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "",
		},
		{
			name:      "Service error",
			inputBody: `{"match":"12345","reward":10.5,"reward_type":"percent"}`,
			inputArgs: args{
				match:      "12345",
				reward:     10.5,
				rewardType: "percent",
			},
			mockBehavior: func(r *mock_handler.MockService, args args) {
				r.EXPECT().CreateProductReward(args.match, args.reward, args.rewardType).Return(errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_handler.NewMockService(c)
			tt.mockBehavior(service, tt.inputArgs)

			handler := Handler{service: service}

			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/products", bytes.NewBufferString(tt.inputBody))

			// Make Request
			handler.CreateProductReward()(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_RegisterNewOrder(t *testing.T) {
	type args struct {
		order models.Order
	}
	type mockBehavior func(r *mock_handler.MockService, args args)

	tests := []struct {
		name                 string
		inputBody            string
		inputArgs            args
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:      "Ok",
			inputBody: `{"order":12345,"goods":[{"description":"product1","price":100.5}]}`,
			inputArgs: args{
				order: models.Order{
					Order: 12345,
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_handler.MockService, args args) {
				r.EXPECT().RegisterNewOrder(args.order).Return(false, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: "",
		},
		{
			name:      "Order already exists",
			inputBody: `{"order":12345,"goods":[{"description":"product1","price":100.5}]}`,
			inputArgs: args{
				order: models.Order{
					Order: 12345,
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_handler.MockService, args args) {
				r.EXPECT().RegisterNewOrder(args.order).Return(true, nil)
			},
			expectedStatusCode:   409,
			expectedResponseBody: "",
		},
		{
			name:                 "Wrong input",
			inputBody:            `{"order":"12345"`,
			inputArgs:            args{},
			mockBehavior:         func(r *mock_handler.MockService, args args) {},
			expectedStatusCode:   400,
			expectedResponseBody: "",
		},
		{
			name:      "Service error",
			inputBody: `{"order":12345,"goods":[{"description":"product1","price":100.5}]}`,
			inputArgs: args{
				order: models.Order{
					Order: 12345,
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_handler.MockService, args args) {
				r.EXPECT().RegisterNewOrder(args.order).Return(false, errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_handler.NewMockService(c)
			tt.mockBehavior(service, tt.inputArgs)

			handler := Handler{service: service}

			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/orders", bytes.NewBufferString(tt.inputBody))

			// Make Request
			handler.RegisterNewOrder()(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.Equal(t, tt.expectedResponseBody, w.Body.String())
		})
	}
}

func TestHandler_GetAccrualInfo(t *testing.T) {
	type args struct {
		order int64
	}
	type mockBehavior func(r *mock_handler.MockService, args args, status string, accrual float64, exist bool)

	tests := []struct {
		name                 string
		orderNumber          string
		inputArgs            args
		status               string
		accrual              float64
		exist                bool
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Ok",
			orderNumber: "12345",
			inputArgs: args{
				order: 12345,
			},
			status:  models.PROCESSED,
			accrual: 100.5,
			exist:   true,
			mockBehavior: func(r *mock_handler.MockService, args args, status string, accrual float64, exist bool) {
				r.EXPECT().GetAccrualInfo(args.order).Return(status, accrual, exist, nil)
			},
			expectedStatusCode:   200,
			expectedResponseBody: `{"order":12345,"status":"PROCESSED","accrual":100.5}`,
		},
		{
			name:        "Order not found",
			orderNumber: "12345",
			inputArgs: args{
				order: 12345,
			},
			status:  "",
			accrual: 0,
			exist:   false,
			mockBehavior: func(r *mock_handler.MockService, args args, status string, accrual float64, exist bool) {
				r.EXPECT().GetAccrualInfo(args.order).Return(status, accrual, exist, nil)
			},
			expectedStatusCode:   404,
			expectedResponseBody: "",
		},
		{
			name:                 "Wrong input",
			orderNumber:          "abc",
			inputArgs:            args{},
			mockBehavior:         func(r *mock_handler.MockService, args args, status string, accrual float64, exist bool) {},
			expectedStatusCode:   400,
			expectedResponseBody: "",
		},
		{
			name:        "Service error",
			orderNumber: "12345",
			inputArgs: args{
				order: 12345,
			},
			mockBehavior: func(r *mock_handler.MockService, args args, status string, accrual float64, exist bool) {
				r.EXPECT().GetAccrualInfo(args.order).Return("", 0.0, false, errors.New("service error"))
			},
			expectedStatusCode:   500,
			expectedResponseBody: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			service := mock_handler.NewMockService(c)
			tt.mockBehavior(service, tt.inputArgs, tt.status, tt.accrual, tt.exist)

			handler := Handler{service: service}

			// Create Request
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/orders/accrual?number="+tt.orderNumber, nil)

			// Make Request
			handler.GetAccrualInfo()(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.expectedStatusCode == 200 {
				var response models.AccrualInfo
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.inputArgs.order, response.Order)
				assert.Equal(t, tt.status, response.Status)
				assert.Equal(t, tt.accrual, response.Accrual)
			} else {
				assert.Equal(t, tt.expectedResponseBody, w.Body.String())
			}
		})
	}
}
