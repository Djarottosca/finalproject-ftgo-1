package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var ErrProductNotFound = errors.New("produk tidak ditemukan")

// cacheTTL (60 detik)
const cacheTTL = 60 * time.Second

type Service struct {
	repo  Repository
	cache Cache
}

func NewService(repo Repository, cache Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

// List menerapkan cache-aside: cek cache dulu, kalau miss baru query
func (s *Service) List(ctx context.Context, req ListRequest) (*ListResponse, error) {
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

func (s *Service) Detail(ctx context.Context, slug string) (*ProductDetailResponse, error) {
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
