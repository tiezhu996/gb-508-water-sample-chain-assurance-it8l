package repository

import (
	"context"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"gorm.io/gorm"
)

// ResultReviewRepository owns all persistence operations for 结果复核.
type ResultReviewRepository interface {
	List(context.Context, dto.PageQuery) (Page[model.ResultReview], error)
	Get(context.Context, uint) (model.ResultReview, error)
	Create(context.Context, *model.ResultReview) error
	Update(context.Context, uint, uint, *model.ResultReview) error
	Delete(context.Context, uint) error
	CountByStatus(context.Context) (map[string]int64, error)
}

type resultReviewRepository struct {
	store *Store[model.ResultReview]
}

func NewResultReviewRepository(db *gorm.DB) ResultReviewRepository {
	return &resultReviewRepository{store: NewStore[model.ResultReview](db)}
}

func (r *resultReviewRepository) List(ctx context.Context, q dto.PageQuery) (Page[model.ResultReview], error) {
	return r.store.List(ctx, q)
}
func (r *resultReviewRepository) Get(ctx context.Context, id uint) (model.ResultReview, error) {
	return r.store.Get(ctx, id)
}
func (r *resultReviewRepository) Create(ctx context.Context, item *model.ResultReview) error {
	return r.store.Create(ctx, item)
}
func (r *resultReviewRepository) Update(ctx context.Context, id, version uint, item *model.ResultReview) error {
	return r.store.Update(ctx, id, version, item)
}
func (r *resultReviewRepository) Delete(ctx context.Context, id uint) error {
	return r.store.Delete(ctx, id)
}
func (r *resultReviewRepository) CountByStatus(ctx context.Context) (map[string]int64, error) {
	return r.store.CountByStatus(ctx)
}
