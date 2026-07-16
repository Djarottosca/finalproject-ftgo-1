package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/slug"
)

var (
	ErrProductNotFound = errors.New("produk tidak ditemukan")
	ErrForbidden       = errors.New("produk bukan milik supplier ini")
	ErrInvalidDiscount = errors.New("discount_type harus percentage, fixed, atau kosong")
	ErrInvalidStock    = errors.New("stok hasil akhir tidak boleh negatif")
)

// cacheTTL (60 detik)
const cacheTTL = 60 * time.Second

// Service defines the product use cases exposed to the handler layer.
type Service interface {
	List(ctx context.Context, req ListRequest) (*ListResponse, error)
	Detail(ctx context.Context, slug string) (*ProductDetailResponse, error)
	Create(ctx context.Context, supplierID uint64, req CreateProductRequest) (*ProductDetailResponse, error)
	ListMine(ctx context.Context, supplierID uint64) ([]ProductResponse, error)
	Update(ctx context.Context, supplierID, productID uint64, req UpdateProductRequest) (*ProductDetailResponse, error)
	SetDiscount(ctx context.Context, supplierID, productID uint64, req DiscountRequest) (*ProductDetailResponse, error)
	AdjustStock(ctx context.Context, supplierID, productID uint64, req StockAdjustRequest) (*ProductDetailResponse, error)
	Delete(ctx context.Context, supplierID, productID uint64) error
}

type service struct {
	repo  Repository
	cache Cache
}

// NewService returns the Service implementation.
func NewService(repo Repository, cache Cache) Service {
	return &service{repo: repo, cache: cache}
}

// List menerapkan cache-aside: cek cache dulu, kalau miss baru query
func (s *service) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
	cacheKey := fmt.Sprintf("products:list:kw=%s:cat=%d:page=%d:limit=%d", req.Keyword, req.CategoryID, req.Page, req.Limit)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var resp ListResponse
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
			return &resp, nil
		}
		//gagal unmarshal (format cache lama berubah), abaikan cache dan lanjut query database seperti biasa

	}

	filter := ListFilter{
		Keyword:    req.Keyword,
		CategoryID: req.CategoryID,
		Page:       req.Page,
		Limit:      req.Limit,
	}

	products, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("query produk: %w", err)
	}

	items := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, toProductResponse(p))
	}

	resp := &ListResponse{
		Items:      items,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalItems: total,
	}

	if payload, jsonErr := json.Marshal(resp); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKey, string(payload), cacheTTL)
	}

	return resp, nil
}

func (s *service) Detail(ctx context.Context, slug string) (*ProductDetailResponse, error) {
	cacheKey := fmt.Sprintf("products:detail:%s", slug)

	if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
		var resp ProductDetailResponse
		if jsonErr := json.Unmarshal([]byte(cached), &resp); jsonErr == nil {
			return &resp, nil
		}
	}

	p, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("query produk: %w", err)
	}

	resp := toProductDetailResponse(*p)

	if payload, jsonErr := json.Marshal(resp); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKey, string(payload), cacheTTL)
	}

	return &resp, nil
}

// Create adds a new product owned by supplierID.
func (s *service) Create(ctx context.Context, supplierID uint64, req CreateProductRequest) (*ProductDetailResponse, error) {
	p := &models.Product{
		ProductName: req.ProductName,
		ProductSlug: slug.Generate(req.ProductName),
		CategoryID:  req.CategoryID,
		Unit:        req.Unit,
		Stock:       req.Stock,
		SupplierID:  supplierID,
		Price:       req.Price,
		Description: req.Description,
		Status:      models.ProductStatusActive,
	}
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("buat produk: %w", err)
	}

	resp := toProductDetailResponse(*p)
	return &resp, nil
}

// ListMine returns every product owned by supplierID, active or not.
func (s *service) ListMine(ctx context.Context, supplierID uint64) ([]ProductResponse, error) {
	products, err := s.repo.FindAllBySupplier(ctx, supplierID)
	if err != nil {
		return nil, fmt.Errorf("query produk supplier: %w", err)
	}

	items := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, toProductResponse(p))
	}
	return items, nil
}

// findOwned loads a product by ID and checks it belongs to supplierID.
func (s *service) findOwned(ctx context.Context, supplierID, productID uint64) (*models.Product, error) {
	p, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if p.SupplierID != supplierID {
		return nil, ErrForbidden
	}
	return p, nil
}

// Update overwrites the editable fields of a product owned by supplierID.
func (s *service) Update(ctx context.Context, supplierID, productID uint64, req UpdateProductRequest) (*ProductDetailResponse, error) {
	p, err := s.findOwned(ctx, supplierID, productID)
	if err != nil {
		return nil, err
	}

	p.ProductName = req.ProductName
	p.CategoryID = req.CategoryID
	p.Unit = req.Unit
	p.Price = req.Price
	p.Description = req.Description

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("ubah produk: %w", err)
	}

	resp := toProductDetailResponse(*p)
	return &resp, nil
}

// SetDiscount sets or clears the discount on a product owned by supplierID.
func (s *service) SetDiscount(ctx context.Context, supplierID, productID uint64, req DiscountRequest) (*ProductDetailResponse, error) {
	p, err := s.findOwned(ctx, supplierID, productID)
	if err != nil {
		return nil, err
	}

	switch req.DiscountType {
	case "":
		p.DiscountType = nil
		p.DiscountAmount = nil
	case models.ProductDiscountPercentage, models.ProductDiscountFixed:
		if req.DiscountType == models.ProductDiscountPercentage && req.DiscountAmount > 100 {
			return nil, ErrInvalidDiscount
		}
		discountType := req.DiscountType
		amount := req.DiscountAmount
		p.DiscountType = &discountType
		p.DiscountAmount = &amount
	default:
		return nil, ErrInvalidDiscount
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("ubah diskon: %w", err)
	}

	resp := toProductDetailResponse(*p)
	return &resp, nil
}

// AdjustStock adds req.Delta to the current stock of a product owned by
// supplierID. Delta can be negative to deduct stock.
func (s *service) AdjustStock(ctx context.Context, supplierID, productID uint64, req StockAdjustRequest) (*ProductDetailResponse, error) {
	p, err := s.findOwned(ctx, supplierID, productID)
	if err != nil {
		return nil, err
	}

	newStock := p.Stock + req.Delta
	if newStock < 0 {
		return nil, ErrInvalidStock
	}
	p.Stock = newStock

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("ubah stok: %w", err)
	}

	resp := toProductDetailResponse(*p)
	return &resp, nil
}

// Delete removes a product owned by supplierID.
func (s *service) Delete(ctx context.Context, supplierID, productID uint64) error {
	if _, err := s.findOwned(ctx, supplierID, productID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, productID)
}

func toProductResponse(p models.Product) ProductResponse {
	return ProductResponse{
		ID:             p.ID,
		ProductName:    p.ProductName,
		ProductSlug:    p.ProductSlug,
		CategoryID:     p.CategoryID,
		CategoryName:   p.Category.CategoryName,
		Unit:           p.Unit,
		Stock:          p.Stock,
		Price:          p.Price,
		DiscountType:   p.DiscountType,
		DiscountAmount: p.DiscountAmount,
		FinalPrice:     p.FinalPrice(),
		Status:         p.Status,
	}
}

func toProductDetailResponse(p models.Product) ProductDetailResponse {
	images := make([]string, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, img.ImageURL)
	}

	return ProductDetailResponse{
		ProductResponse: toProductResponse(p),
		Description:     p.Description,
		SupplierID:      p.SupplierID,
		Images:          images,
	}
}
