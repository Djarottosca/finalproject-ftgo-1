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

// --- Sales report ---

// SalesSummary is the revenue rollup, counting only PAID orders (money that
// actually landed; pending/expired funds still sit with the payment provider).
type SalesSummary struct {
	TotalRevenue  float64 `json:"total_revenue"`
	PaidOrders    int     `json:"paid_orders"`
	AvgOrderValue float64 `json:"avg_order_value"`
}

// StatusCount is how many orders sit in a given payment status.
type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// SalesReportResponse combines the revenue rollup with a breakdown of order
// counts per payment status, so admin sees both the money and the funnel.
type SalesReportResponse struct {
	Summary        SalesSummary  `json:"summary"`
	OrdersByStatus []StatusCount `json:"orders_by_status"`
}

// --- Transaction monitoring ---

// TransactionItem is one order line with its payment status, for the
// read-only transaction monitor.
type TransactionItem struct {
	OrderID       int     `json:"order_id"`
	UserID        int     `json:"user_id"`
	FinalPrice    float64 `json:"final_price"`
	OrderStatus   string  `json:"order_status"`
	PaymentStatus string  `json:"payment_status"`
	PaymentRef    string  `json:"payment_reference,omitempty"`
}
