package repository

import (
	"context"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"gorm.io/gorm"
)

// AssayMethodRepository owns all persistence operations for 检测方法.
type AssayMethodRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.AssayMethod], error)
	Get(context.Context, uint) (model.AssayMethod, error)
	Create(context.Context, *model.AssayMethod) error
	Update(context.Context, uint, uint, *model.AssayMethod) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
}

type assayMethodRepository struct {
	store *Store[model.AssayMethod]
}

func NewAssayMethodRepository(db *gorm.DB) AssayMethodRepository {
	return &assayMethodRepository{store: NewStore[model.AssayMethod](db)}
}

func (r *assayMethodRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.AssayMethod], error) {
	return r.store.List(ctx, q)
}
func (r *assayMethodRepository) Get(ctx context.Context, id uint) (model.AssayMethod, error) {
	return r.store.Get(ctx, id)
}
func (r *assayMethodRepository) Create(ctx context.Context, item *model.AssayMethod) error {
	return r.store.Create(ctx, item)
}
func (r *assayMethodRepository) Update(ctx context.Context, id, version uint, item *model.AssayMethod) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *assayMethodRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *assayMethodRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
