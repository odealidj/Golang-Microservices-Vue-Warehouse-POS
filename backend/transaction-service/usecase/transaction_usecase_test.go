package usecase

import (
	"context"
	"micro-warehouse/transaction-service/model"
	"micro-warehouse/transaction-service/pkg/httpclient"
	"testing"
)

type stubTransactionRepo struct {
	saved *model.Transaction
}

func (s *stubTransactionRepo) GetDashboardStats(ctx context.Context) (int64, int64, int64, error) {
	return 0, 0, 0, nil
}
func (s *stubTransactionRepo) GetDashboardStatsByMerchant(ctx context.Context, merchantID uint) (int64, int64, int64, error) {
	return 0, 0, 0, nil
}
func (s *stubTransactionRepo) GetTransactions(ctx context.Context, page, limit int, search, sortBy, sortOrder string, merchantID uint) ([]model.Transaction, int64, error) {
	return nil, 0, nil
}
func (s *stubTransactionRepo) GetTransactionByID(ctx context.Context, id uint) (*model.Transaction, error) {
	return nil, nil
}
func (s *stubTransactionRepo) CreateTransaction(ctx context.Context, transaction model.Transaction) (int64, error) {
	copyTx := transaction
	s.saved = &copyTx
	return 99, nil
}
func (s *stubTransactionRepo) UpdatePaymentStatus(ctx context.Context, orderID string, paymentStatus, paymentMethod, transactionID, fraudStatus string) error {
	return nil
}

type stubMerchantClient struct {
	stocks map[uint]int
}

func (s *stubMerchantClient) GetMerchantsByKeeperID(ctx context.Context, keeperID uint) ([]httpclient.Merchant, error) {
	return nil, nil
}
func (s *stubMerchantClient) GetMerchantByID(ctx context.Context, merchantID uint) (*httpclient.Merchant, error) {
	return &httpclient.Merchant{ID: merchantID, Name: "Toko Sembako Jaya"}, nil
}
func (s *stubMerchantClient) GetMerchantProducts(ctx context.Context, merchantID uint) ([]httpclient.MerchantProduct, error) {
	return nil, nil
}
func (s *stubMerchantClient) GetMerchantProductStock(ctx context.Context, merchantID uint, productID uint) (*httpclient.MerchantProduct, error) {
	return &httpclient.MerchantProduct{
		MerchantID:  merchantID,
		ProductID:   productID,
		ProductName: "Beras Premium 5kg",
		Stock:       s.stocks[productID],
	}, nil
}

type stubProductClient struct {
	products map[uint]httpclient.ProductResponse
}

func (s *stubProductClient) GetProductByID(ctx context.Context, productID uint) (*httpclient.ProductResponse, error) {
	p := s.products[productID]
	return &p, nil
}
func (s *stubProductClient) GetProductByBarcode(ctx context.Context, barcode string) (*httpclient.ProductResponse, error) {
	return nil, nil
}
func (s *stubProductClient) GetProducts(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]httpclient.ProductResponse, error) {
	return nil, nil
}
func (s *stubProductClient) HealthCheck(ctx context.Context) error {
	return nil
}

type stubUserClient struct{}

func (s *stubUserClient) GetUserByID(ctx context.Context, userID uint) (*httpclient.UserResponse, error) {
	return &httpclient.UserResponse{ID: userID, RoleName: "Keeper"}, nil
}
func (s *stubUserClient) HealthCheck(ctx context.Context) error {
	return nil
}

func TestCreateTransaction_ServerSidePricingAndPPN11(t *testing.T) {
	repo := &stubTransactionRepo{}
	merchantClient := &stubMerchantClient{stocks: map[uint]int{1: 10}}
	productClient := &stubProductClient{
		products: map[uint]httpclient.ProductResponse{
			1: {ID: 1, Name: "Beras Premium 5kg", Price: 75000},
		},
	}

	uc := NewTransactionUsecase(repo, merchantClient, nil, productClient, &stubUserClient{})

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
	if repo.saved == nil {
		t.Fatal("expected transaction to be saved")
	}
	if repo.saved.SubTotal != 150000 {
		t.Fatalf("expected SubTotal 150000, got %d", repo.saved.SubTotal)
	}
	if repo.saved.TaxTotal != 16500 {
		t.Fatalf("expected 11%% PPN 16500, got %d", repo.saved.TaxTotal)
	}
	if repo.saved.GrandTotal != 166500 {
		t.Fatalf("expected GrandTotal 166500, got %d", repo.saved.GrandTotal)
	}
}

func TestCreateTransaction_InsufficientStock(t *testing.T) {
	repo := &stubTransactionRepo{}
	merchantClient := &stubMerchantClient{stocks: map[uint]int{1: 1}}
	productClient := &stubProductClient{
		products: map[uint]httpclient.ProductResponse{
			1: {ID: 1, Name: "Minyak Goreng 2L", Price: 38000},
		},
	}

	uc := NewTransactionUsecase(repo, merchantClient, nil, productClient, &stubUserClient{})

	tx := &model.Transaction{
		Name:       "Siti",
		MerchantID: 1,
		TransactionProducts: []model.TransactionProduct{
			{ProductID: 1, Quantity: 3},
		},
	}

	_, err := uc.CreateTransaction(context.Background(), tx)
	if err == nil {
		t.Fatal("expected error when quantity exceeds merchant stock")
	}
	if repo.saved != nil {
		t.Fatal("expected repository not to be called when stock validation fails")
	}
}

func TestCreateTransaction_EmptyProducts(t *testing.T) {
	uc := NewTransactionUsecase(&stubTransactionRepo{}, &stubMerchantClient{}, nil, &stubProductClient{}, &stubUserClient{})
	if _, err := uc.CreateTransaction(context.Background(), &model.Transaction{}); err == nil {
		t.Fatal("expected error for empty transaction products")
	}
}
