package cart

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

type Repository interface {
	// Upsert adds qty if (user_id, product_id) already exists in the cart,
	// otherwise inserts a new row.
	Upsert(ctx context.Context, userID int, productID int, qty int) error
	UpdateQty(ctx context.Context, userID int, productID int, qty int) (int64, error)
	Delete(ctx context.Context, userID int, productID int) (int64, error)
	FindAllByUser(ctx context.Context, userID int) ([]models.Cart, error)
}

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Upsert(ctx context.Context, userID int, productID int, qty int) error {
	cart := models.Cart{UserID: userID, ProductID: productID, Qty: qty}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"qty":        gorm.Expr("carts.qty + ?", qty),
			"updated_at": gorm.Expr("now()"),
		}),
	}).Create(&cart).Error
}

func (r *gormRepository) UpdateQty(ctx context.Context, userID int, productID int, qty int) (int64, error) {
	result := r.db.WithContext(ctx).Model(&models.Cart{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Update("qty", qty)
	return result.RowsAffected, result.Error
}

func (r *gormRepository) Delete(ctx context.Context, userID int, productID int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Delete(&models.Cart{})
	return result.RowsAffected, result.Error
}

func (r *gormRepository) FindAllByUser(ctx context.Context, userID int) ([]models.Cart, error) {
	var carts []models.Cart
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&carts).Error
	if err != nil {
		return nil, err
	}
	return carts, nil
}
