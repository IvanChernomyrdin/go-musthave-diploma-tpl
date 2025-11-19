package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"go-musthave-diploma-tpl/internal/accrual/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// MockService is a mock implementation of the Service interface
type MockService struct {
	CreateProductRewardFunc func(match string, reward float64, rewardType string) error
}

func (m *MockService) CreateProductReward(match string, reward float64, rewardType string) error {
	if m.CreateProductRewardFunc != nil {
		return m.CreateProductRewardFunc(match, reward, rewardType)
	}
	return nil
}

func TestHandler_CreateProductReward(t *testing.T) {
	tests := []struct {
		name           string
		productReward  models.ProductReward
		serviceError   error
		expectedStatus int
	}{
		{
			name: "Successful creation",
			productReward: models.ProductReward{
				Match:      "product123",
				Reward:     10.5,
				RewardType: "percentage",
			},
			serviceError:   nil,
			expectedStatus: http.StatusOK,
		},
		{
			name: "Duplicate product reward",
			productReward: models.ProductReward{
				Match:      "product123",
				Reward:     10.5,
				RewardType: "percentage",
			},
			serviceError:   storage.ErrKeyExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "Internal server error",
			productReward: models.ProductReward{
				Match:      "product123",
				Reward:     10.5,
				RewardType: "percentage",
			},
			serviceError:   errors.New("internal error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock service
			mockService := &MockService{
				CreateProductRewardFunc: func(match string, reward float64, rewardType string) error {
					return tt.serviceError
				},
			}

			// Create handler with mock service
			handler := NewHandler(mockService)

			// Create request with JSON body
			productRewardJSON, _ := json.Marshal(tt.productReward)
			req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBuffer(productRewardJSON))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler function
			handler.CreateProductReward()(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}
		})
	}

	// Test invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		// Create a mock service
		mockService := &MockService{}

		// Create handler with mock service
		handler := NewHandler(mockService)

		// Create request with invalid JSON body
		req := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler function
		handler.CreateProductReward()(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})

	// Test request body read error
	t.Run("Request body read error", func(t *testing.T) {
		// Create a mock service
		mockService := &MockService{}

		// Create handler with mock service
		handler := NewHandler(mockService)

		// Create request with a faulty body reader
		req := httptest.NewRequest(http.MethodPost, "/api/products", &faultyReader{})
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call the handler function
		handler.CreateProductReward()(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}

// faultyReader is a faulty io.Reader that always returns an error
type faultyReader struct{}

func (f *faultyReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

// Close method to satisfy io.ReadCloser interface
func (f *faultyReader) Close() error {
	return nil
}

// Test that faultyReader implements io.ReadCloser
var _ io.ReadCloser = (*faultyReader)(nil)
