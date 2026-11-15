package usecase

import (
	"context"
	"errors"
	"micro-warehouse/product-service/model"
	"testing"
)

type stubProductRepo struct {
	products   map[uint]*model.Product
	byBarcode  map[string]*model.Product
	lastCreate *model.Product
	lastUpdate *model.Product
	err        error
}

func (s *stubProductRepo) CreateProduct(ctx context.Context, product *model.Product) error {
	if s.err != nil {
		return s.err
	}
	s.lastCreate = product
	if s.products == nil {
		s.products = map[uint]*model.Product{}
	}
	s.products[product.ID] = product
	return nil
}

func (s *stubProductRepo) GetAllProducts(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]model.Product, int64, error) {
	if s.err != nil {
		return nil, 0, s.err
	}
	var out []model.Product
	for _, p := range s.products {
		out = append(out, *p)
	}
	return out, int64(len(out)), nil
}

func (s *stubProductRepo) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	if s.err != nil {
		return nil, s.err
	}
	p, ok := s.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func (s *stubProductRepo) GetProductByBarcode(ctx context.Context, barcode string) (*model.Product, error) {
	if s.err != nil {
		return nil, s.err
	}
	p, ok := s.byBarcode[barcode]
	if !ok {
		return nil, errors.New("barcode not found")
	}
	return p, nil
}

func (s *stubProductRepo) UpdateProduct(ctx context.Context, product *model.Product) error {
	if s.err != nil {
		return s.err
	}
	s.lastUpdate = product
	return nil
}

func (s *stubProductRepo) DeleteProduct(ctx context.Context, id uint) error {
	return s.err
}

func TestProductUsecase_CRUDAndBarcodeLookup(t *testing.T) {
	p1 := &model.Product{ID: 10, Name: "Minyak Goreng 2L", Barcode: "8991002003004", Price: 36000}
	repo := &stubProductRepo{
		products:  map[uint]*model.Product{},
		byBarcode: map[string]*model.Product{p1.Barcode: p1},
	}
	uc := NewProductUsecase(repo)

	if err := uc.CreateProduct(context.Background(), p1); err != nil {
		t.Fatalf("CreateProduct failed: %v", err)
	}
	if repo.lastCreate == nil || repo.lastCreate.Name != "Minyak Goreng 2L" {
		t.Fatalf("expected created product to be recorded, got %+v", repo.lastCreate)
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
	if repo.lastUpdate == nil || repo.lastUpdate.Price != 38000 {
		t.Fatalf("expected updated price 38000, got %+v", repo.lastUpdate)
	}
}
