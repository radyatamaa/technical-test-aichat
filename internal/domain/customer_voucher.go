package domain

import (
	"context"
	"gorm.io/gorm"
)

type CustomerVoucher struct {
	ID        int            `gorm:"column:id;primarykey;autoIncrement:true"`
	CustomerID  *int `gorm:"type:bigint(20);column:customer_id"`
	Customer               Customer       `gorm:"foreignkey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;->"`
	VoucherCode string `gorm:"type:varchar(255);column:voucher_code"`
	IsRedeem bool `gorm:"bool;column:is_redeem"`
}




// TableName name of table
func (r CustomerVoucher) TableName() string {
	return "customer_voucher"
}

// MysqlCustomerVoucherRepository Repository Interface
type MysqlCustomerVoucherRepository interface {
	SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error
	FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error)
	Update(ctx context.Context, data CustomerVoucher) error
	UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error
	UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error
	Store(ctx context.Context, data CustomerVoucher) (CustomerVoucher, error)
	StoreWithTx(ctx context.Context, tx *gorm.DB, data CustomerVoucher) (int, error)
	Delete(ctx context.Context, id int) (int, error)
	SoftDelete(ctx context.Context, id int) (int, error)
	DB() *gorm.DB
}
