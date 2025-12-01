package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/accrual/models"
	mock_service "go-musthave-diploma-tpl/internal/accrual/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestService_CreateProductReward(t *testing.T) {
	type args struct {
		match      string
		reward     float64
		rewardType string
	}
	type mockBehavior func(r *mock_service.MockRepository, args args)

	tests := []struct {
		name         string
		inputArgs    args
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name: "Ok",
			inputArgs: args{
				match:      "12345",
				reward:     10.5,
				rewardType: "%",
			},
			mockBehavior: func(r *mock_service.MockRepository, args args) {
				r.EXPECT().CreateProductReward(context.Background(), args.match, args.reward, args.rewardType).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Repository error",
			inputArgs: args{
				match:      "12345",
				reward:     10.5,
				rewardType: "%",
			},
			mockBehavior: func(r *mock_service.MockRepository, args args) {
				r.EXPECT().CreateProductReward(context.Background(), args.match, args.reward, args.rewardType).Return(errors.New("repository error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockRepository(c)
			tt.mockBehavior(repo, tt.inputArgs)

			service := NewService(repo)

			// Test
			err := service.CreateProductReward(context.Background(), tt.inputArgs.match, tt.inputArgs.reward, tt.inputArgs.rewardType)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RegisterNewOrder(t *testing.T) {
	type args struct {
		order models.Order
	}
	type mockBehavior func(r *mock_service.MockRepository, args args, exist bool)

	tests := []struct {
		name         string
		inputArgs    args
		mockBehavior mockBehavior
		exist        bool
		expectedErr  bool
	}{
		{
			name: "Ok - New Order",
			inputArgs: args{
				order: models.Order{
					Order: "12345",
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, exist bool) {
				r.EXPECT().CheckOrderExists(int64(12345)).Return(exist, nil)
				r.EXPECT().RegisterNewOrder(context.Background(), int64(12345), args.order.Goods, models.Registered).Return(nil)
			},
			exist:       false,
			expectedErr: false,
		},
		{
			name: "Order already exists",
			inputArgs: args{
				order: models.Order{
					Order: "12345",
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, exist bool) {
				r.EXPECT().CheckOrderExists(int64(12345)).Return(exist, nil)
				// RegisterNewOrder should not be called when order already exists
			},
			exist:       true,
			expectedErr: false,
		},
		{
			name: "Invalid order number",
			inputArgs: args{
				order: models.Order{
					Order: "abc",
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, exist bool) {},
			exist:        false,
			expectedErr:  true,
		},
		{
			name: "CheckOrderExists error",
			inputArgs: args{
				order: models.Order{
					Order: "12345",
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, exist bool) {
				r.EXPECT().CheckOrderExists(int64(12345)).Return(false, errors.New("database error"))
			},
			exist:       false,
			expectedErr: true,
		},
		{
			name: "RegisterNewOrder error",
			inputArgs: args{
				order: models.Order{
					Order: "12345",
					Goods: []models.Goods{
						{
							Description: "product1",
							Price:       100.5,
						},
					},
				},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, exist bool) {
				r.EXPECT().CheckOrderExists(int64(12345)).Return(false, nil)
				r.EXPECT().RegisterNewOrder(context.Background(), int64(12345), args.order.Goods, models.Registered).Return(errors.New("database error"))
			},
			exist:       false,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockRepository(c)
			tt.mockBehavior(repo, tt.inputArgs, tt.exist)

			service := NewService(repo)

			// Test
			exist, err := service.RegisterNewOrder(context.Background(), tt.inputArgs.order)

			assert.Equal(t, tt.exist, exist)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetAccrualInfo(t *testing.T) {
	type args struct {
		order int64
	}
	type mockBehavior func(r *mock_service.MockRepository, args args, status string, accrual float64, exist bool)

	tests := []struct {
		name         string
		inputArgs    args
		status       string
		accrual      float64
		exist        bool
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name: "Ok - Order exists",
			inputArgs: args{
				order: 12345,
			},
			status:  models.Processed,
			accrual: 100.5,
			exist:   true,
			mockBehavior: func(r *mock_service.MockRepository, args args, status string, accrual float64, exist bool) {
				r.EXPECT().CheckOrderExists(args.order).Return(exist, nil)
				r.EXPECT().GetAccrualInfo(args.order).Return(status, accrual, nil)
			},
			expectedErr: false,
		},
		{
			name: "Order not found",
			inputArgs: args{
				order: 12345,
			},
			status:  "",
			accrual: 0.0,
			exist:   false,
			mockBehavior: func(r *mock_service.MockRepository, args args, status string, accrual float64, exist bool) {
				r.EXPECT().CheckOrderExists(args.order).Return(exist, nil)
			},
			expectedErr: false,
		},
		{
			name: "CheckOrderExists error",
			inputArgs: args{
				order: 12345,
			},
			mockBehavior: func(r *mock_service.MockRepository, args args, status string, accrual float64, exist bool) {
				r.EXPECT().CheckOrderExists(args.order).Return(false, errors.New("database error"))
			},
			expectedErr: true,
		},
		{
			name: "GetAccrualInfo error",
			inputArgs: args{
				order: 12345,
			},
			exist: true,
			mockBehavior: func(r *mock_service.MockRepository, args args, status string, accrual float64, exist bool) {
				r.EXPECT().CheckOrderExists(args.order).Return(exist, nil)
				r.EXPECT().GetAccrualInfo(args.order).Return("", 0.0, errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockRepository(c)
			tt.mockBehavior(repo, tt.inputArgs, tt.status, tt.accrual, tt.exist)

			service := NewService(repo)

			// Test
			status, accrual, exist, err := service.GetAccrualInfo(tt.inputArgs.order)

			assert.Equal(t, tt.exist, exist)
			if tt.exist {
				assert.Equal(t, tt.status, status)
				assert.Equal(t, tt.accrual, accrual)
			}

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_processOrders(t *testing.T) {
	type args struct {
		products []models.ProductReward
		orders   []models.ParseMatch
	}
	type mockBehavior func(r *mock_service.MockRepository, args args)

	tests := []struct {
		name         string
		inputArgs    args
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name: "GetProductsInfo error",
			inputArgs: args{
				products: []models.ProductReward{},
				orders:   []models.ParseMatch{},
			},
			mockBehavior: func(r *mock_service.MockRepository, args args) {
				r.EXPECT().GetProductsInfo().Return(nil, errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockRepository(c)
			tt.mockBehavior(repo, tt.inputArgs)

			service := NewService(repo)

			// Test
			logger := zap.NewNop().Sugar()
			err := service.processOrders(context.Background(), logger)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_updateOrderAccrual(t *testing.T) {
	type args struct {
		orderID    int64
		orderItems []models.ParseMatch
		product    models.ProductReward
	}
	type mockBehavior func(r *mock_service.MockRepository, args args, totalAccrual float64)

	tests := []struct {
		name         string
		inputArgs    args
		totalAccrual float64
		mockBehavior mockBehavior
		expectedErr  bool
	}{
		{
			name: "Ok - Percentage reward",
			inputArgs: args{
				orderID: 12345,
				orderItems: []models.ParseMatch{
					{
						Order: 12345,
						Price: 100,
					},
				},
				product: models.ProductReward{
					Match:      "12345",
					Reward:     10,
					RewardType: "pt",
				},
			},
			totalAccrual: 10.0,
			mockBehavior: func(r *mock_service.MockRepository, args args, totalAccrual float64) {
				r.EXPECT().UpdateAccrualInfo(gomock.Any(), args.orderID, totalAccrual, models.Processed).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Ok - Absolute reward",
			inputArgs: args{
				orderID: 12345,
				orderItems: []models.ParseMatch{
					{
						Order: 12345,
						Price: 100,
					},
				},
				product: models.ProductReward{
					Match:      "12345",
					Reward:     50,
					RewardType: "abs",
				},
			},
			totalAccrual: 50.0,
			mockBehavior: func(r *mock_service.MockRepository, args args, totalAccrual float64) {
				r.EXPECT().UpdateAccrualInfo(gomock.Any(), args.orderID, totalAccrual, models.Processed).Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "UpdateAccrualInfo error",
			inputArgs: args{
				orderID: 12345,
				orderItems: []models.ParseMatch{
					{
						Order: 12345,
						Price: 100,
					},
				},
				product: models.ProductReward{
					Match:      "12345",
					Reward:     10,
					RewardType: "pt",
				},
			},
			totalAccrual: 10.0,
			mockBehavior: func(r *mock_service.MockRepository, args args, totalAccrual float64) {
				r.EXPECT().UpdateAccrualInfo(gomock.Any(), args.orderID, totalAccrual, models.Processed).Return(errors.New("database error"))
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Init Dependencies
			c := gomock.NewController(t)
			defer c.Finish()

			repo := mock_service.NewMockRepository(c)
			tt.mockBehavior(repo, tt.inputArgs, tt.totalAccrual)

			service := NewService(repo)

			// Test
			logger := zap.NewNop().Sugar()
			err := service.updateOrderAccrual(context.Background(), logger, tt.inputArgs.orderID, tt.inputArgs.orderItems, tt.inputArgs.product)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Listener(t *testing.T) {

	assert.NotPanics(t, func() {
		c := gomock.NewController(t)
		defer c.Finish()

		repo := mock_service.NewMockRepository(c)
		service := NewService(repo)

		ctx, cancel := context.WithCancel(context.Background())
		logger := zap.NewNop().Sugar()

		cancel()

		service.Listener(ctx, logger, time.Millisecond)
	})
}
