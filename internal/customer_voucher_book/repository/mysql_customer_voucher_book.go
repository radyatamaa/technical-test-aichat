package repository

import (
	"context"
	"github.com/radyatamaa/go-cqrs-microservices/api_gateway_service/internal/domain"
	"github.com/radyatamaa/go-cqrs-microservices/pkg/database/paginator"
	"github.com/radyatamaa/go-cqrs-microservices/pkg/zaplogger"
	"gorm.io/gorm"
	"strings"
)

type mysqlCustomerVoucherBookRepository struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}



func NewMysqlCCustomerVoucherBookRepository(db *gorm.DB, zapLogger zaplogger.Logger) domain.MysqlCustomerVoucherBookRepository {
	return &mysqlCustomerVoucherBookRepository{
		db:        db,
		zapLogger: zapLogger,
	}
}


func (c mysqlCustomerVoucherBookRepository) DB() *gorm.DB {
	return c.db
}

func (c mysqlCustomerVoucherBookRepository) FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args).Select(strings.Join(fields, ",")).Error; err != nil {
		return nil, err
	}
	return model,nil
}

func (c mysqlCustomerVoucherBookRepository) SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error {

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

func (c mysqlCustomerVoucherBookRepository) Update(ctx context.Context, data domain.CustomerVoucherBook) error {

	err := c.db.WithContext(ctx).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c mysqlCustomerVoucherBookRepository) UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error {

	return c.db.WithContext(ctx).Table(domain.CustomerVoucherBook{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlCustomerVoucherBookRepository) Store(ctx context.Context, data domain.CustomerVoucherBook) (domain.CustomerVoucherBook, error) {

	err := c.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (c mysqlCustomerVoucherBookRepository) Delete(ctx context.Context, id int) (int, error) {

	err := c.db.WithContext(ctx).Exec("delete from "+domain.CustomerVoucherBook{}.TableName()+" where id =?", id).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlCustomerVoucherBookRepository) SoftDelete(ctx context.Context, id int) (int, error) {
	var data domain.CustomerVoucherBook

	err := c.db.WithContext(ctx).Where("id = ?", id).Delete(&data).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlCustomerVoucherBookRepository) UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error {

	return tx.WithContext(ctx).Table(domain.CustomerVoucherBook{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlCustomerVoucherBookRepository) StoreWithTx(ctx context.Context, tx *gorm.DB, data domain.CustomerVoucherBook) (int, error) {

	err := tx.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data.ID, err
	}
	return data.ID, nil
}


