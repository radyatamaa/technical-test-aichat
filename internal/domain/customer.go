package domain

import (
	"context"
	"database/sql"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/go-cqrs-microservices/pkg/helper"
	"gorm.io/gorm"
	"mime/multipart"
	"time"
)

type Customer struct {
	ID        int            `gorm:"column:id;primarykey;autoIncrement:true"`
	FirstName    string         `gorm:"type:varchar(255);column:first_name"`
	LastName     string         `gorm:"type:varchar(255);column:last_name"`
	Gender     string         `gorm:"type:varchar(50);column:gender"`
	DateOfBirth     string         `gorm:"type:date;column:date_of_birth"`
	ContactNumber     string         `gorm:"type:varchar(50);column:contact_number"`
	Email      string         `gorm:"type:varchar(255);column:email"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
}

// TableName name of table
func (r Customer) TableName() string {
	return "customers"
}

// CustomerUseCase UseCase Interface
type CustomerUseCase interface {
	VerifyPhotoCustomer(beegoCtx *beegoContext.Context, customerId int,file *multipart.FileHeader) (*CustomerVerifyPhotoResponse,error)
	GetVoucherByCustomerId(beegoCtx *beegoContext.Context, customerId int) (*CustomerVoucherBookResponse,error)
}

// MysqlCustomerRepository Repository Interface
type MysqlCustomerRepository interface {
	SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error
	FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error)
	Update(ctx context.Context, data Customer) error
	UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error
	UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error
	Store(ctx context.Context, data Customer) (Customer, error)
	StoreWithTx(ctx context.Context, tx *gorm.DB, data Customer) (int, error)
	Delete(ctx context.Context, id int) (int, error)
	SoftDelete(ctx context.Context, id int) (int, error)
	DB() *gorm.DB
}

func SeederData(db *gorm.DB)  {
	dataCustomer := make([]Customer,1500)
	dataPurchaseTransaction := make([]PurchaseTransaction,1500)
	for i := range dataCustomer {
		dataCustomer[i] = Customer{
			ID:            0,
			FirstName:     "Customer First Name" + helper.IntToString(i+1),
			LastName:      "Customer last Name" + helper.IntToString(i+1),
			Gender:        "Laki-Laki",
			DateOfBirth:   "1999-15-06",
			ContactNumber: "081572345351",
			Email:         helper.RandomString(10),
		}

		dataPurchaseTransaction[i] = PurchaseTransaction{
			ID:            0,
			CustomerID:    sql.NullInt32{Int32: int32(i + 1) ,Valid: true},
			Customer:      Customer{},
			TotalSpent:    100,
			TotalSaving:   50,
			TransactionAt: time.Now(),
		}
	}

	db.Create(&dataCustomer)

	db.Create(&dataPurchaseTransaction)


	dataCustomerVoucher := make([]CustomerVoucher,1000)
	for i := range dataCustomerVoucher {
		dataCustomerVoucher[i] = CustomerVoucher{
			ID:          0,
			CustomerID:  sql.NullInt32{},
			Customer:    Customer{},
			VoucherCode: helper.RandomString(10),
			IsRedeem:    false,
		}
	}

	db.Create(dataCustomerVoucher)
}