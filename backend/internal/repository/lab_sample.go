package repository

import (
	"context"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"gorm.io/gorm"
)

// LabSampleRepository owns all persistence operations for 实验室样本.
type LabSampleRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.LabSample], error)
	Get(context.Context, uint) (model.LabSample, error)
	Create(context.Context, *model.LabSample) error
	Update(context.Context, uint, uint, *model.LabSample) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
}

type labSampleRepository struct {
	store *Store[model.LabSample]
}

func NewLabSampleRepository(db *gorm.DB) LabSampleRepository {
	return &labSampleRepository{store: NewStore[model.LabSample](db)}
}

func (r *labSampleRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.LabSample], error) {
	return r.store.List(ctx, q)
}
func (r *labSampleRepository) Get(ctx context.Context, id uint) (model.LabSample, error) {
	return r.store.Get(ctx, id)
}
func (r *labSampleRepository) Create(ctx context.Context, item *model.LabSample) error {
	return r.store.Create(ctx, item)
}
func (r *labSampleRepository) Update(ctx context.Context, id, version uint, item *model.LabSample) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *labSampleRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *labSampleRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
