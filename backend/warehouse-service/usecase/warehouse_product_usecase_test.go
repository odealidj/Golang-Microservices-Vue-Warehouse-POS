package usecase

import (
	"context"
	"micro-warehouse/warehouse-service/model"
	"micro-warehouse/warehouse-service/pkg/httpclient"
	"testing"

	"gorm.io/gorm"
)

type stubWarehouseProductRepo struct {
	existing      *model.WarehouseProduct
	lookupErr     error
	createCalled  bool
	updateCalled  bool
	lastUpdated   *model.WarehouseProduct
	detailWh      *model.Warehouse
}

func (s *stubWarehouseProductRepo) GetDetailWarehouse(ctx context.Context, warehouseID uint) (*model.Warehouse, error) {
	return s.detailWh, nil
}
func (s *stubWarehouseProductRepo) GetDetailWarehouseProductByID(ctx context.Context, warehouseProductID uint) (*model.WarehouseProduct, error) {
	return nil, nil
}
func (s *stubWarehouseProductRepo) CreateWarehouseProduct(ctx context.Context, warehouseProduct *model.WarehouseProduct) error {
	s.createCalled = true
	return nil
}
func (s *stubWarehouseProductRepo) GetWarehouseProductByWarehouseIDAndProductID(ctx context.Context, warehouseID, productID uint) (*model.WarehouseProduct, error) {
	return s.existing, s.lookupErr
}
func (s *stubWarehouseProductRepo) UpdateWarehouseProduct(ctx context.Context, warehouseProduct *model.WarehouseProduct) error {
	s.updateCalled = true
	s.lastUpdated = warehouseProduct
	return nil
}
func (s *stubWarehouseProductRepo) DeleteWarehouseProduct(ctx context.Context, warehouseProductID uint) error {
	return nil
}
func (s *stubWarehouseProductRepo) DeleteAllWarehouseProductByProductID(ctx context.Context, productID uint) error {
	return nil
}
func (s *stubWarehouseProductRepo) GetWarehouseProductByProductID(ctx context.Context, productID uint) ([]model.WarehouseProduct, error) {
	return nil, nil
}
func (s *stubWarehouseProductRepo) GetProductTotalStock(ctx context.Context, productID uint) (int, error) {
	return 0, nil
}

type stubWarehouseProductClient struct {
	products map[uint]httpclient.ProductResponse
}

func (s *stubWarehouseProductClient) GetProductByID(ctx context.Context, productID uint) (*httpclient.ProductResponse, error) {
	p := s.products[productID]
	return &p, nil
}
func (s *stubWarehouseProductClient) GetProducts(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]httpclient.ProductResponse, error) {
	return nil, nil
}
func (s *stubWarehouseProductClient) HealthCheck(ctx context.Context) error { return nil }

func TestCreateWarehouseProduct_CreatesWhenNotFound(t *testing.T) {
	repo := &stubWarehouseProductRepo{lookupErr: gorm.ErrRecordNotFound}
	uc := NewWarehouseProductUsecase(repo, &stubWarehouseProductClient{})

	err := uc.CreateWarehouseProduct(context.Background(), &model.WarehouseProduct{
		WarehouseID: 1,
		ProductID:   2,
		Stock:       50,
	})
	if err != nil {
		t.Fatalf("CreateWarehouseProduct failed: %v", err)
	}
	if !repo.createCalled || repo.updateCalled {
		t.Fatalf("expected createCalled=true updateCalled=false, got create=%v update=%v", repo.createCalled, repo.updateCalled)
	}
}

func TestCreateWarehouseProduct_UpdatesWhenExisting(t *testing.T) {
	repo := &stubWarehouseProductRepo{
		existing: &model.WarehouseProduct{ID: 7, WarehouseID: 1, ProductID: 2, Stock: 20},
	}
	uc := NewWarehouseProductUsecase(repo, &stubWarehouseProductClient{})

	wp := &model.WarehouseProduct{WarehouseID: 1, ProductID: 2, Stock: 75}
	if err := uc.CreateWarehouseProduct(context.Background(), wp); err != nil {
		t.Fatalf("CreateWarehouseProduct failed: %v", err)
	}
	if !repo.updateCalled || repo.createCalled {
		t.Fatalf("expected updateCalled=true createCalled=false, got create=%v update=%v", repo.createCalled, repo.updateCalled)
	}
	if repo.lastUpdated == nil || repo.lastUpdated.ID != 7 || repo.lastUpdated.Stock != 75 {
		t.Fatalf("expected existing ID=7 and Stock=75, got %+v", repo.lastUpdated)
	}
}

func TestGetDetailWarehouse_EnrichesProductsFromClient(t *testing.T) {
	repo := &stubWarehouseProductRepo{
		detailWh: &model.Warehouse{
			ID:   1,
			Name: "Gudang Pusat",
			WarehouseProducts: []model.WarehouseProduct{
				{ID: 11, WarehouseID: 1, ProductID: 101, Stock: 30},
			},
		},
	}
	client := &stubWarehouseProductClient{
		products: map[uint]httpclient.ProductResponse{
			101: {ID: 101, Name: "Gula Pasir 1kg", Price: 17500},
		},
	}
	uc := NewWarehouseProductUsecase(repo, client)

	wh, products, err := uc.GetDetailWarehouse(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDetailWarehouse failed: %v", err)
	}
	if wh.Name != "Gudang Pusat" || len(products) != 1 || products[0].Name != "Gula Pasir 1kg" {
		t.Fatalf("unexpected detail result: wh=%+v products=%+v", wh, products)
	}
}
