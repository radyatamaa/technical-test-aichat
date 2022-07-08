package repository

import (
	"context"
	"github.com/radyatamaa/go-cqrs-microservices/api_gateway_service/internal/domain"
	"github.com/radyatamaa/go-cqrs-microservices/pkg/database/paginator"
	"github.com/radyatamaa/go-cqrs-microservices/pkg/zaplogger"
	"gorm.io/gorm"
	"strings"
)

type mysqlCustomerRepository struct {
	zapLogger zaplogger.Logger
	db        *gorm.DB
}



func NewMysqlCustomerRepository(db *gorm.DB, zapLogger zaplogger.Logger) domain.MysqlCustomerRepository {
	return &mysqlCustomerRepository{
		db:        db,
		zapLogger: zapLogger,
	}
}

func (c mysqlCustomerRepository) DB() *gorm.DB {
	return c.db
}

func (c mysqlCustomerRepository) FetchWithFilter(ctx context.Context, limit int, offset int, order string, fields, associate, filter []string, model interface{}, args ...interface{}) (interface{}, error) {
	p := paginator.NewPaginator(c.db, offset, limit, model)
	if err := p.FindWithFilter(ctx, order, fields, associate, filter, args).Select(strings.Join(fields, ",")).Error; err != nil {
		return nil, err
	}
	return model,nil
}

func (c mysqlCustomerRepository) SingleWithFilter(ctx context.Context, fields, associate, filter []string, model interface{}, args ...interface{}) error {

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

func (c mysqlCustomerRepository) Update(ctx context.Context, data domain.Customer) error {

	err := c.db.WithContext(ctx).Updates(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (c mysqlCustomerRepository) UpdateSelectedField(ctx context.Context, field []string, values map[string]interface{}, id int) error {

	return c.db.WithContext(ctx).Table(domain.Customer{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlCustomerRepository) Store(ctx context.Context, data domain.Customer) (domain.Customer, error) {

	err := c.db.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (c mysqlCustomerRepository) Delete(ctx context.Context, id int) (int, error) {

	err := c.db.WithContext(ctx).Exec("delete from "+domain.Customer{}.TableName()+" where id =?", id).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlCustomerRepository) SoftDelete(ctx context.Context, id int) (int, error) {
	var data domain.Customer

	err := c.db.WithContext(ctx).Where("id = ?", id).Delete(&data).Error
	if err != nil {
		return id, err
	}
	return id, nil
}

func (c mysqlCustomerRepository) UpdateSelectedFieldWithTx(ctx context.Context, tx *gorm.DB, field []string, values map[string]interface{}, id int) error {

	return tx.WithContext(ctx).Table(domain.Customer{}.TableName()).Select(field).Where("id =?", id).Updates(values).Error
}

func (c mysqlCustomerRepository) StoreWithTx(ctx context.Context, tx *gorm.DB, data domain.Customer) (int, error) {

	err := tx.WithContext(ctx).Create(&data).Error
	if err != nil {
		return data.ID, err
	}
	return data.ID, nil
}


