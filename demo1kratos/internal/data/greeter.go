package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/orzkratos/demokratos/demo1kratos/internal/biz"
	"github.com/orzkratos/demokratos/demo1kratos/internal/pkg/models"
	"gorm.io/gorm"
)

type greeterRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
	return &greeterRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *greeterRepo) Save(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	db := r.data.DB()

	// Use GORM transaction to save record
	// 使用 GORM 事务保存记录
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := &models.Record{
			Message: g.Hello,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, errors.New(500, "DB_ERROR", err.Error())
	}
	return g, nil
}

func (r *greeterRepo) Update(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	return g, nil
}

func (r *greeterRepo) FindByID(ctx context.Context, id int64) (*biz.Greeter, error) {
	db := r.data.DB()

	var record models.Record
	if err := db.WithContext(ctx).First(&record, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(404, "RECORD_NOT_FOUND", err.Error())
		}
		return nil, errors.New(500, "DB_ERROR", err.Error())
	}

	return &biz.Greeter{
		Hello: record.Message,
	}, nil
}

func (r *greeterRepo) ListByHello(context.Context, string) ([]*biz.Greeter, error) {
	return nil, nil
}

func (r *greeterRepo) ListAll(context.Context) ([]*biz.Greeter, error) {
	return nil, nil
}
