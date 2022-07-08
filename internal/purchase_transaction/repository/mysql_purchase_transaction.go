package repository

import (
	"context"
	"strings"

	"github.com/radyatamaa/technical-test-aichat/internal/domain"
	"github.com/radyatamaa/technical-test-aichat/pkg/database/paginator"
	"github.com/radyatamaa/technical-test-aichat/pkg/zaplogger"
	"gorm.io/gorm"
)

type mysqlPurchaseTransactionRepository struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}

func NewPurchaseTransactionRepository(db *gorm.DB, zapLogger zaplogger.Logger) domain.MysqlPurchaseTransactionRepository {
	return &mysqlPurchaseTransactionRepository{
		db:        db,
		zapLogger: zapLogger,
	}
}

func (c mysqlPurchaseTransactionRepository) DB() *gorm.DB {
	return c.db
}

func (c mysqlPurchaseTransactionRepository) CountFilter(ctx context.Context, associate []string, model interface{}, criteria []string, args ...interface{}) (int, error) {
	var count int64
	db := c.db.WithContext(ctx)

	if len(associate) > 0 {
		for _, v := range associate {
			db.Joins(v)
		}
	}

	if len(criteria) > 0 && len(args) == len(criteria) {
		for i := range criteria {
			db = db.Where(criteria[i], args[i])
		}
	}

	if err := db.Model(model).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func (c mysqlPurchaseTransactionRepository) FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args).Select(strings.Join(fields, ",")).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (c mysqlPurchaseTransactionRepository) SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error {

	db := c.db.WithContext(ctx)

	if len(fields) > 0 {
		db = db.Select(strings.Join(fields, ","))
	}
	if len(associate) > 0 {
		for _, v := range associate {
			db.Joins(v)
		}
	}

	if err := db.First(model, strings.Join(filter, ","), args).Error; err != nil {
		return err
	}

	return nil
}

func (c mysqlPurchaseTransactionRepository) Update(ctx context.Context, data domain.PurchaseTransaction) error {

	err := c.db.WithContext(ctx).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c mysqlPurchaseTransactionRepository) UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error {

	return c.db.WithContext(ctx).Table(domain.PurchaseTransaction{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlPurchaseTransactionRepository) Store(ctx context.Context, data domain.PurchaseTransaction) (domain.PurchaseTransaction, error) {

	err := c.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (c mysqlPurchaseTransactionRepository) Delete(ctx context.Context, id int) (int, error) {

	err := c.db.WithContext(ctx).Exec("delete from "+domain.PurchaseTransaction{}.TableName()+" where id =?", id).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlPurchaseTransactionRepository) SoftDelete(ctx context.Context, id int) (int, error) {
	var data domain.PurchaseTransaction

	err := c.db.WithContext(ctx).Where("id = ?", id).Delete(&data).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlPurchaseTransactionRepository) UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error {

	return tx.WithContext(ctx).Table(domain.PurchaseTransaction{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlPurchaseTransactionRepository) StoreWithTx(ctx context.Context, tx *gorm.DB, data domain.PurchaseTransaction) (int, error) {

	err := tx.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data.ID, err
	}
	return data.ID, nil
}
