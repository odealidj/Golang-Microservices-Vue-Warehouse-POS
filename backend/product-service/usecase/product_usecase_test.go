package usecase

import (
	"context"
	"testing"

	"micro-warehouse/product-service/mocks"
	"micro-warehouse/product-service/model"

	"go.uber.org/mock/gomock"
)

func TestProductUsecase_CRUDAndBarcodeLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockProductRepositoryInterface(ctrl)
	uc := NewProductUsecase(repo)

	p1 := &model.Product{ID: 10, Name: "Minyak Goreng 2L", Barcode: "8991002003004", Price: 36000}

	repo.EXPECT().CreateProduct(gomock.Any(), gomock.AssignableToTypeOf(&model.Product{})).
		DoAndReturn(func(_ context.Context, product *model.Product) error {
			if product.Name != "Minyak Goreng 2L" {
				t.Fatalf("expected product name Minyak Goreng 2L, got %q", product.Name)
			}
			return nil
		})
	repo.EXPECT().GetProductByID(gomock.Any(), uint(10)).Return(p1, nil)
	repo.EXPECT().GetProductByBarcode(gomock.Any(), p1.Barcode).Return(p1, nil)
	repo.EXPECT().GetAllProducts(gomock.Any(), 1, 10, "", "id", "asc").Return([]model.Product{*p1}, int64(1), nil)
	repo.EXPECT().UpdateProduct(gomock.Any(), gomock.AssignableToTypeOf(&model.Product{})).
		DoAndReturn(func(_ context.Context, product *model.Product) error {
			if product.Price != 38000 {
				t.Fatalf("expected updated price 38000, got %v", product.Price)
			}
			return nil
		})

	if err := uc.CreateProduct(context.Background(), p1); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}

	gotByID, err := uc.GetProductByID(context.Background(), 10)
	if err != nil || gotByID.Barcode != p1.Barcode {
		t.Fatalf("GetProductByID failed: got=%+v err=%v", gotByID, err)
	}

	gotByBarcode, err := uc.GetProductByBarcode(context.Background(), p1.Barcode)
	if err != nil || gotByBarcode.ID != 10 {
		t.Fatalf("GetProductByBarcode failed: got=%+v err=%v", gotByBarcode, err)
	}

	all, total, err := uc.GetAllProducts(context.Background(), 1, 10, "", "id", "asc")
	if err != nil || total != 1 || len(all) != 1 {
		t.Fatalf("GetAllProducts unexpected result: len=%d total=%d err=%v", len(all), total, err)
	}

	p1.Price = 38000
	if err := uc.UpdateProduct(context.Background(), p1); err != nil {
		t.Fatalf("UpdateProduct failed: %v", err)
	}
}
