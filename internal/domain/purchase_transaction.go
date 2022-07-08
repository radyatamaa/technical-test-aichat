package domain

import (
	"context"
	"database/sql"
	"gorm.io/gorm"
	"time"
)

type PurchaseTransaction struct {
	ID        int            `gorm:"column:id;primarykey;autoIncrement:true"`
	CustomerID  sql.NullInt32 `gorm:"type:int;column:customer_id"`
	//Customer               Customer       `gorm:"foreignkey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	TotalSpent     float64         `gorm:"type:decimal(10,2);column:total_spent"`
	TotalSaving     float64         `gorm:"type:decimal(10,2);column:total_saving"`
	TransactionAt time.Time      `gorm:"column:transaction_at"`
}

// TableName name of table
func (r PurchaseTransaction) TableName() string {
	return "purchase_transactions"
}

// MysqlPurchaseTransactionRepository Repository Interface
type MysqlPurchaseTransactionRepository interface {
	CountFilter(ctx context.Context, associate []string, model interface{},criteria []string, args ...interface{}) (int, error)
	SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error
	FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error)
	Update(ctx context.Context, data PurchaseTransaction) error
	UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error
	UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error
	Store(ctx context.Context, data PurchaseTransaction) (PurchaseTransaction, error)
	StoreWithTx(ctx context.Context, tx *gorm.DB, data PurchaseTransaction) (int, error)
	Delete(ctx context.Context, id int) (int, error)
	SoftDelete(ctx context.Context, id int) (int, error)
	DB() *gorm.DB
}
