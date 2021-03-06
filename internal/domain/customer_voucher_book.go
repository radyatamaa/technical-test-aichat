package domain

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type CustomerVoucherBook struct {
	ID        int            `gorm:"column:id;primarykey;autoIncrement:true"`
	CustomerID  int `gorm:"type:bigint(20);column:customer_id"`
	Customer               Customer       `gorm:"foreignkey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	CustomerVoucherID int `gorm:"type:bigint(20);column:customer_voucher_id"`
	CustomerVoucher               CustomerVoucher       `gorm:"foreignkey:CustomerVoucherID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	ExpiredDate 	time.Time `gorm:"column:expired_date"`
}

// TableName name of table
func (r CustomerVoucherBook) TableName() string {
	return "customer_voucher_books"
}

// MysqlCustomerVoucherBookRepository Repository Interface
type MysqlCustomerVoucherBookRepository interface {
	SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error
	FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error)
	Update(ctx context.Context, data CustomerVoucherBook) error
	UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error
	UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error
	Store(ctx context.Context, data CustomerVoucherBook) (CustomerVoucherBook, error)
	StoreWithTx(ctx context.Context, tx *gorm.DB, data CustomerVoucherBook) (int, error)
	Delete(ctx context.Context, id int) (int, error)
	SoftDelete(ctx context.Context, id int) (int, error)
	DB() *gorm.DB
}