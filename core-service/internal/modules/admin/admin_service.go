package admin

import "context"

// defaultLowStockThreshold is the fallback when the request doesn't specify
// one. Overridable per-request via query param; not a config value because it
// changes rarely and the query param already covers the dynamic case.
const defaultLowStockThreshold = 10

// Service defines the admin reporting use cases exposed to the handler layer.
type Service interface {
	StockReport(ctx context.Context, threshold int) (*StockReportResponse, error)
	SalesReport(ctx context.Context) (*SalesReportResponse, error)
	Transactions(ctx context.Context) ([]TransactionItem, error)
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) StockReport(ctx context.Context, threshold int) (*StockReportResponse, error) {
	if threshold <= 0 {
		threshold = defaultLowStockThreshold
	}

	products, err := s.repo.ListProductsByStock(ctx)
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.StockSummary(ctx, threshold)
	if err != nil {
		return nil, err
	}

	return toStockReport(products, summary, threshold), nil
}

func (s *service) SalesReport(ctx context.Context) (*SalesReportResponse, error) {
	summary, err := s.repo.SalesSummary(ctx)
	if err != nil {
		return nil, err
	}

	byStatus, err := s.repo.OrdersByStatus(ctx)
	if err != nil {
		return nil, err
	}

	return &SalesReportResponse{
		Summary:        summary,
		OrdersByStatus: byStatus,
	}, nil
}

func (s *service) Transactions(ctx context.Context) ([]TransactionItem, error) {
	return s.repo.ListTransactions(ctx)
}
