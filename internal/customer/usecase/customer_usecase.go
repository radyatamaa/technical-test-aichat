package usecase

import (
	"context"
	"mime/multipart"
	"strings"
	"time"

	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/radyatamaa/technical-test-aichat/internal/domain"
	"github.com/radyatamaa/technical-test-aichat/pkg/helper"
	"github.com/radyatamaa/technical-test-aichat/pkg/response"
	"github.com/radyatamaa/technical-test-aichat/pkg/zaplogger"
	"gorm.io/gorm"
)

type customerUseCase struct {
	zapLogger                          zaplogger.Logger
	contextTimeout                     time.Duration
	mysqlCustomerRepository            domain.MysqlCustomerRepository
	mysqlCustomerVoucherRepository     domain.MysqlCustomerVoucherRepository
	mysqlCustomerVoucherBookRepository domain.MysqlCustomerVoucherBookRepository
	mysqlPurchaseTransactionRepository domain.MysqlPurchaseTransactionRepository
}

func NewCustomerUseCase(timeout time.Duration,
	mysqlCustomerRepository domain.MysqlCustomerRepository,
	mysqlCustomerVoucherRepository domain.MysqlCustomerVoucherRepository,
	mysqlCustomerVoucherBookRepository domain.MysqlCustomerVoucherBookRepository,
	mysqlPurchaseTransactionRepository domain.MysqlPurchaseTransactionRepository,
	zapLogger zaplogger.Logger) domain.CustomerUseCase {
	return &customerUseCase{
		mysqlCustomerRepository:            mysqlCustomerRepository,
		mysqlCustomerVoucherRepository:     mysqlCustomerVoucherRepository,
		mysqlPurchaseTransactionRepository: mysqlPurchaseTransactionRepository,
		contextTimeout:                     timeout,
		zapLogger:                          zapLogger,
		mysqlCustomerVoucherBookRepository: mysqlCustomerVoucherBookRepository,
	}
}

// QUERY CUSTOMER
func (r customerUseCase) singleCustomerWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.Customer, error) {
	var entity domain.Customer
	if err := r.mysqlCustomerRepository.SingleWithFilter(
		ctx,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

// QUERY CUSTOMER VOUCHER BOOK
func (r customerUseCase) singleCustomerVoucherBookWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.CustomerVoucherBook, error) {
	var entity domain.CustomerVoucherBook
	if err := r.mysqlCustomerVoucherBookRepository.SingleWithFilter(
		ctx,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

// QUERY CUSTOMER VOUCHER
func (r customerUseCase) fetchCustomerVoucherWithFilter(ctx context.Context, limit, offset int, filter []string, args ...interface{}) ([]domain.CustomerVoucher, error) {

	if purchaseTransaction, err := r.mysqlCustomerVoucherRepository.FetchWithFilter(
		ctx,
		limit,
		offset,
		"RAND()",
		[]string{
			"*",
		},
		[]string{},
		filter,
		&[]domain.CustomerVoucher{}, args); err != nil {
		return nil, err
	} else {
		if result, ok := purchaseTransaction.(*[]domain.CustomerVoucher); !ok {
			return []domain.CustomerVoucher{}, nil
		} else {
			return *result, nil
		}
	}
}

func (r customerUseCase) singleCustomerVoucherWithFilter(ctx context.Context, filter []string, args ...interface{}) (*domain.CustomerVoucher, error) {
	var entity domain.CustomerVoucher
	if err := r.mysqlCustomerVoucherRepository.SingleWithFilter(
		ctx,
		[]string{
			"*",
		},
		[]string{},
		filter,
		&entity, args...); err != nil {
		return nil, err
	}
	return &entity, nil
}

// QUERY PURCHASE TRANSACTION
func (r customerUseCase) summaryPurchaseTransactionWithFilter(ctx context.Context, filter []string, args ...interface{}) ([]domain.PurchaseTransaction, error) {

	if purchaseTransaction, err := r.mysqlPurchaseTransactionRepository.FetchWithFilter(
		ctx,
		1,
		0,
		"transaction_at DESC",
		[]string{
			"SUM(total_spent) AS total_spent",
			"SUM(total_saving) AS total_saving",
		},
		[]string{},
		filter,
		&[]domain.PurchaseTransaction{}, args); err != nil {
		//beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		//r.zapLogger.SetMessageLog2(err)
		return nil, err
	} else {
		if result, ok := purchaseTransaction.(*[]domain.PurchaseTransaction); !ok {
			return []domain.PurchaseTransaction{}, nil
		} else {
			return *result, nil
		}
	}
}

func (r customerUseCase) countPurchaseTransactionWithFilter(ctx context.Context, filter []string, args ...interface{}) (int, error) {
	var entity domain.PurchaseTransaction
	var result int
	result, err := r.mysqlPurchaseTransactionRepository.CountFilter(
		ctx,
		[]string{},
		&entity,
		filter,
		args...)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r customerUseCase) VerifyPhotoCustomer(beegoCtx *beegoContext.Context, customerId int, file *multipart.FileHeader) (*domain.CustomerVerifyPhotoResponse, error) {
	c, cancel := context.WithTimeout(beegoCtx.Request.Context(), r.contextTimeout)
	defer cancel()

	first, err := r.singleCustomerVoucherWithFilter(c, []string{"customer_id"}, customerId)
	if err != nil && err != gorm.ErrRecordNotFound{
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if first != nil{
		return nil, response.ErrCustomerAlreadyGetVoucher
	}

	voucherBookCheckCustomer, err := r.singleCustomerVoucherBookWithFilter(c,
		[]string{
			"customer_id"},
		customerId)
	if err != nil && err != gorm.ErrRecordNotFound {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if voucherBookCheckCustomer == nil {
		return nil, response.ErrCustomerNotYetBookVoucher
	}

	if time.Now().After(voucherBookCheckCustomer.ExpiredDate) {
		return nil, response.ErrCustomerBookVoucherExpired
	}

	//VALIDATE IMAGE BY SIZE
	sizeKb := float64(file.Size / 1024)
	if !strings.Contains(file.Filename, "face") || (sizeKb < 50) {
		return nil, response.ErrCustomerVerifyImage
	}



	err = r.mysqlCustomerVoucherRepository.UpdateSelectedField(c,
		[]string{"customer_id", "is_redeem"},
		map[string]interface{}{
			"customer_id": customerId,
			"is_redeem":   true,
		},
		voucherBookCheckCustomer.CustomerVoucherID,
	)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	first, err = r.singleCustomerVoucherWithFilter(c, []string{"id"}, voucherBookCheckCustomer.CustomerVoucherID)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}
	return &domain.CustomerVerifyPhotoResponse{VoucherCode: first.VoucherCode}, nil

}

func (r customerUseCase) GetVoucherByCustomerId(beegoCtx *beegoContext.Context, customerId int) (*domain.CustomerVoucherBookResponse, error) {
	c, cancel := context.WithTimeout(beegoCtx.Request.Context(), r.contextTimeout)
	defer cancel()

	first, err := r.singleCustomerWithFilter(c, []string{"id =?"}, customerId)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	// VALIDATION CUSTOMER ALREADY GET VOUCHER
	firstCV, err := r.singleCustomerVoucherWithFilter(c, []string{"customer_id =?"}, customerId)
	if err != nil && err != gorm.ErrRecordNotFound {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if firstCV != nil {
		return nil, response.ErrCustomerAlreadyGetVoucher
	}

	// VALIDATION MIN 3 COMPLETE TRANSACTION
	startDate := time.Now()
	endDate := startDate.AddDate(0, 0, 30)

	countPurchaseTransaction, err := r.countPurchaseTransactionWithFilter(c,
		[]string{"customer_id", "DATE(transaction_at) >= DATE(?)", "DATE(transaction_at) <= DATE(?)"},
		first.ID,
		startDate,
		endDate)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if countPurchaseTransaction < 3 {
		return nil, response.ErrTransactionCompletePurchase30Days
	}

	// VALIDATION TRANSACTION MIN 100$
	summary, err := r.summaryPurchaseTransactionWithFilter(c, []string{"customer_id = ?"}, first.ID)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if len(summary) == 0 {
		return nil, response.ErrTransactionMinimum
	}

	if summary[0].TotalSpent < 100 {
		return nil, response.ErrTransactionMinimum
	}

	// VALIDATION ALREADY BOOK VOUCHER
	voucherBookCheckCustomer, err := r.singleCustomerVoucherBookWithFilter(c,
		[]string{
			"customer_id",
			"expired_date > ?"},
		customerId,
		time.Now())
	if err != nil && err != gorm.ErrRecordNotFound {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	if voucherBookCheckCustomer != nil {
		return nil, response.ErrCustomerAlreadyBookVoucher
	}

	fetchCV, err := r.fetchCustomerVoucherWithFilter(c, 1000, 0, []string{"is_redeem = ?"}, false)
	if err != nil {
		beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
		return nil, err
	}

	customerVoucherId := 0
	expiredDate := time.Now().Add(time.Minute * 10)

	for i := range fetchCV {
		voucherBook, err := r.singleCustomerVoucherBookWithFilter(c, []string{"customer_voucher_id = ?"}, fetchCV[i].ID)
		if err != nil && err != gorm.ErrRecordNotFound {
			beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
			return nil, err
		}

		if voucherBook != nil {
			if time.Now().After(voucherBook.ExpiredDate) {
				customerVoucherId = fetchCV[i].ID
				_, err = r.mysqlCustomerVoucherBookRepository.Store(c, domain.CustomerVoucherBook{
					CustomerID:        first.ID,
					CustomerVoucherID: fetchCV[i].ID,
					ExpiredDate:       expiredDate,
				})
				if err != nil {
					beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
					return nil, err
				}
			} else {
				continue
			}

		} else {
			customerVoucherId = fetchCV[i].ID
			_, err = r.mysqlCustomerVoucherBookRepository.Store(c, domain.CustomerVoucherBook{
				CustomerID:        first.ID,
				CustomerVoucherID: fetchCV[i].ID,
				ExpiredDate:       expiredDate,
			})
			if err != nil {
				beegoCtx.Input.SetData("stackTrace", r.zapLogger.SetMessageLog(err))
				return nil, err
			}
		}

		break
	}

	if customerVoucherId == 0 {
		return nil, response.ErrVoucherNotAvailable
	}

	return &domain.CustomerVoucherBookResponse{Expired: expiredDate.Format(helper.DateTimeFormatDefault)}, nil
}
