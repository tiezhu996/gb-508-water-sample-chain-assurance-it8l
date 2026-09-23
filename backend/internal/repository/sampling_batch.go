package repository

import (
	"context"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"gorm.io/gorm"
)

// SamplingBatchRepository owns all persistence operations for 采样批次.
type SamplingBatchRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.SamplingBatch], error)
	Get(context.Context, uint) (model.SamplingBatch, error)
	Create(context.Context, *model.SamplingBatch) error
	Update(context.Context, uint, uint, *model.SamplingBatch) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
}

type samplingBatchRepository struct {
	store *Store[model.SamplingBatch]
}

func NewSamplingBatchRepository(db *gorm.DB) SamplingBatchRepository {
	return &samplingBatchRepository{store: NewStore[model.SamplingBatch](db)}
}

func (r *samplingBatchRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.SamplingBatch], error) {
	return r.store.List(ctx, q)
}
func (r *samplingBatchRepository) Get(ctx context.Context, id uint) (model.SamplingBatch, error) {
	return r.store.Get(ctx, id)
}
func (r *samplingBatchRepository) Create(ctx context.Context, item *model.SamplingBatch) error {
	return r.store.Create(ctx, item)
}
func (r *samplingBatchRepository) Update(ctx context.Context, id, version uint, item *model.SamplingBatch) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *samplingBatchRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *samplingBatchRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
