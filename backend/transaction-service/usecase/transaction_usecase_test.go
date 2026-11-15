package usecase

import (
	"context"
	"testing"

	"micro-warehouse/transaction-service/mocks"
	"micro-warehouse/transaction-service/model"
	"micro-warehouse/transaction-service/pkg/httpclient"

	"go.uber.org/mock/gomock"
)

func TestCreateTransaction_ServerSidePricingAndPPN11(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepositoryInterface(ctrl)
	merchantClient := mocks.NewMockMerchantClientInterface(ctrl)
	productClient := mocks.NewMockProductClientInterface(ctrl)
	userClient := mocks.NewMockUserClientInterface(ctrl)

	merchantClient.EXPECT().GetMerchantProductStock(gomock.Any(), uint(1), uint(1)).
		Return(&httpclient.MerchantProduct{
			MerchantID:  1,
			ProductID:   1,
			ProductName: "Beras Premium 5kg",
			Stock:       10,
		}, nil)

	productClient.EXPECT().GetProductByID(gomock.Any(), uint(1)).
		Return(&httpclient.ProductResponse{
			ID:    1,
			Name:  "Beras Premium 5kg",
			Price: 75000,
		}, nil)

	repo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, saved model.Transaction) (int64, error) {
			if saved.SubTotal != 150000 {
				t.Fatalf("expected SubTotal 150000, got %d", saved.SubTotal)
			}
			if saved.TaxTotal != 16500 {
				t.Fatalf("expected 11%% PPN 16500, got %d", saved.TaxTotal)
			}
			if saved.GrandTotal != 166500 {
				t.Fatalf("expected GrandTotal 166500, got %d", saved.GrandTotal)
			}
			return 99, nil
		})

	uc := NewTransactionUsecase(repo, merchantClient, nil, productClient, userClient)

	tx := &model.Transaction{
		Name:       "Budi",
		Phone:      "08123456789",
		Email:      "budi@example.com",
		MerchantID: 1,
		OrderID:    "ORDER_TEST_1",
		TransactionProducts: []model.TransactionProduct{
			{ProductID: 1, Quantity: 2, Price: 1}, // tampered client price should be ignored
		},
	}

	id, err := uc.CreateTransaction(context.Background(), tx)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}
	if id != 99 {
		t.Fatalf("expected id 99, got %d", id)
	}
}

func TestCreateTransaction_InsufficientStock(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepositoryInterface(ctrl)
	merchantClient := mocks.NewMockMerchantClientInterface(ctrl)
	productClient := mocks.NewMockProductClientInterface(ctrl)
	userClient := mocks.NewMockUserClientInterface(ctrl)

	merchantClient.EXPECT().GetMerchantProductStock(gomock.Any(), uint(1), uint(1)).
		Return(&httpclient.MerchantProduct{
			MerchantID:  1,
			ProductID:   1,
			ProductName: "Beras Premium 5kg",
			Stock:       1,
		}, nil)

	repo.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).Times(0)

	uc := NewTransactionUsecase(repo, merchantClient, nil, productClient, userClient)

	tx := &model.Transaction{
		Name:       "Siti",
		MerchantID: 1,
		TransactionProducts: []model.TransactionProduct{
			{ProductID: 1, Quantity: 3},
		},
	}

	if _, err := uc.CreateTransaction(context.Background(), tx); err == nil {
		t.Fatal("expected error when quantity exceeds merchant stock")
	}
}

func TestCreateTransaction_EmptyProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	uc := NewTransactionUsecase(
		mocks.NewMockTransactionRepositoryInterface(ctrl),
		mocks.NewMockMerchantClientInterface(ctrl),
		nil,
		mocks.NewMockProductClientInterface(ctrl),
		mocks.NewMockUserClientInterface(ctrl),
	)
	if _, err := uc.CreateTransaction(context.Background(), &model.Transaction{}); err == nil {
		t.Fatal("expected error for empty transaction products")
	}
}
