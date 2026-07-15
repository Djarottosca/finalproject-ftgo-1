package admin

// StockReportItem is one product line in the stock report, with a low-stock
// flag pre-computed so the client doesn't have to know the threshold.
type StockReportItem struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	SupplierID  int     `json:"supplier_id"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	IsLowStock  bool    `json:"is_low_stock"`
}

// StockSummary is the aggregate rollup across all products.
type StockSummary struct {
	TotalProducts   int   `json:"total_products"`
	TotalStock      int64 `json:"total_stock"`
	LowStockCount   int   `json:"low_stock_count"`
	OutOfStockCount int   `json:"out_of_stock_count"`
}

// StockReportResponse combines the per-product list and the summary rollup,
// so the admin dashboard gets both in one call.
type StockReportResponse struct {
	Threshold int               `json:"threshold"`
	Summary   StockSummary      `json:"summary"`
	Items     []StockReportItem `json:"items"`
}