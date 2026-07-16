package admin

// defaultLowStockThreshold is the fallback when the request doesn't specify
// one. Overridable per-request via query param; not a config value because it
// changes rarely and the query param already covers the dynamic case.
const defaultLowStockThreshold = 10

// Service defines the admin reporting use cases exposed to the handler layer.
type Service interface {
	StockReport(threshold int) (*StockReportResponse, error)
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) StockReport(threshold int) (*StockReportResponse, error) {
	if threshold <= 0 {
		threshold = defaultLowStockThreshold
	}

	products, err := s.repo.ListProductsByStock()
	if err != nil {
		return nil, err
	}

	summary, err := s.repo.StockSummary(threshold)
	if err != nil {
		return nil, err
	}

	return toStockReport(products, summary, threshold), nil
}
