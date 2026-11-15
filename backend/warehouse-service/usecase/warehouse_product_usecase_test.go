package usecase

import (
	"context"
	"testing"

	"micro-warehouse/warehouse-service/mocks"
	"micro-warehouse/warehouse-service/model"
	"micro-warehouse/warehouse-service/pkg/httpclient"

	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestCreateWarehouseProduct_CreatesWhenNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockWarehouseProductRepositoryInterface(ctrl)
	client := mocks.NewMockProductClientInterface(ctrl)

	repo.EXPECT().GetWarehouseProductByWarehouseIDAndProductID(gomock.Any(), uint(1), uint(2)).
		Return(nil, gorm.ErrRecordNotFound)
	repo.EXPECT().CreateWarehouseProduct(gomock.Any(), gomock.AssignableToTypeOf(&model.WarehouseProduct{})).
		Return(nil)
	repo.EXPECT().UpdateWarehouseProduct(gomock.Any(), gomock.Any()).Times(0)

	uc := NewWarehouseProductUsecase(repo, client)
	err := uc.CreateWarehouseProduct(context.Background(), &model.WarehouseProduct{
		WarehouseID: 1,
		ProductID:   2,
		Stock:       50,
	})
	if err != nil {
		t.Fatalf("CreateWarehouseProduct failed: %v", err)
	}
}

func TestCreateWarehouseProduct_UpdatesWhenExisting(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockWarehouseProductRepositoryInterface(ctrl)
	client := mocks.NewMockProductClientInterface(ctrl)

	existing := &model.WarehouseProduct{ID: 7, WarehouseID: 1, ProductID: 2, Stock: 20}
	repo.EXPECT().GetWarehouseProductByWarehouseIDAndProductID(gomock.Any(), uint(1), uint(2)).
		Return(existing, nil)
	repo.EXPECT().UpdateWarehouseProduct(gomock.Any(), gomock.AssignableToTypeOf(&model.WarehouseProduct{})).
		DoAndReturn(func(_ context.Context, wp *model.WarehouseProduct) error {
			if wp.ID != 7 || wp.Stock != 75 {
				t.Fatalf("expected existing ID=7 and Stock=75, got %+v", wp)
			}
			return nil
		})
	repo.EXPECT().CreateWarehouseProduct(gomock.Any(), gomock.Any()).Times(0)

	uc := NewWarehouseProductUsecase(repo, client)
	if err := uc.CreateWarehouseProduct(context.Background(), &model.WarehouseProduct{
		WarehouseID: 1,
		ProductID:   2,
		Stock:       75,
	}); err != nil {
		t.Fatalf("CreateWarehouseProduct failed: %v", err)
	}
}

func TestGetDetailWarehouse_EnrichesProductsFromClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockWarehouseProductRepositoryInterface(ctrl)
	client := mocks.NewMockProductClientInterface(ctrl)

	repo.EXPECT().GetDetailWarehouse(gomock.Any(), uint(1)).
		Return(&model.Warehouse{
			ID:   1,
			Name: "Gudang Pusat",
			WarehouseProducts: []model.WarehouseProduct{
				{ID: 11, WarehouseID: 1, ProductID: 101, Stock: 30},
			},
		}, nil)
	client.EXPECT().GetProductByID(gomock.Any(), uint(101)).
		Return(&httpclient.ProductResponse{ID: 101, Name: "Gula Pasir 1kg", Price: 17500}, nil)

	uc := NewWarehouseProductUsecase(repo, client)
	wh, products, err := uc.GetDetailWarehouse(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetDetailWarehouse failed: %v", err)
	}
	if wh.Name != "Gudang Pusat" || len(products) != 1 || products[0].Name != "Gula Pasir 1kg" {
		t.Fatalf("unexpected detail result: wh=%+v products=%+v", wh, products)
	}
}
