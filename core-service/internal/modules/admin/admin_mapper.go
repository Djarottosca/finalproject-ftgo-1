package admin

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toStockReport(products []models.Product, summary StockSummary, threshold int) *StockReportResponse {
	items := make([]StockReportItem, 0, len(products))
	for i := range products {
		p := products[i]
		items = append(items, StockReportItem{
			ProductID:   p.ID,
			ProductName: p.ProductName,
			SupplierID:  p.SupplierID,
			Stock:       p.Stock,
			Price:       p.Price,
			IsLowStock:  p.Stock < threshold,
		})
	}

	return &StockReportResponse{
		Threshold: threshold,
		Summary:   summary,
		Items:     items,
	}
}
