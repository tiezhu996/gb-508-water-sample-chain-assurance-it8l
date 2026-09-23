package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/constants"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/repository"
)

type SamplingBatchService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.SamplingBatch], error)
	Get(context.Context, uint) (model.SamplingBatch, error)
	Create(context.Context, dto.CreateSamplingBatch, string, string) (model.SamplingBatch, error)
	Update(context.Context, uint, dto.UpdateSamplingBatch, string, string) (model.SamplingBatch, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.SamplingBatch, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
}

type samplingBatchService struct {
	repository repository.SamplingBatchRepository
	security   SecurityService
}

func NewSamplingBatchService(repo repository.SamplingBatchRepository, security SecurityService) SamplingBatchService {
	return &samplingBatchService{repository: repo, security: security}
}

func (s *samplingBatchService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.SamplingBatch], error) {
	return s.repository.List(ctx, query)
}

func (s *samplingBatchService) Get(ctx context.Context, id uint) (model.SamplingBatch, error) {
	return s.repository.Get(ctx, id)
}

func (s *samplingBatchService) Create(ctx context.Context, input dto.CreateSamplingBatch, actor, requestID string) (model.SamplingBatch, error) {
	if err := validateSamplingBatchBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.SamplingBatch{}, err
	}
	item := model.SamplingBatch{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.SamplingBatchInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode: strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.SamplingBatch{}, fmt.Errorf("create 采样批次: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "SamplingBatch", item.ID, "", item.Status, "created 采样批次")
	return item, nil
}

func (s *samplingBatchService) Update(ctx context.Context, id uint, input dto.UpdateSamplingBatch, actor, requestID string) (model.SamplingBatch, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.SamplingBatch{}, err
	}
	if err := validateSamplingBatchBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.SamplingBatch{}, err
	}
	current.Name = strings.TrimSpace(input.Name)
	current.Description = strings.TrimSpace(input.Description)
	current.Facility = strings.TrimSpace(input.Facility)
	current.Owner = strings.TrimSpace(input.Owner)
	current.Category = strings.TrimSpace(input.Category)
	current.RiskLevel = input.RiskLevel
	current.MetricValue = input.MetricValue
	current.MetricUnit = strings.TrimSpace(input.MetricUnit)
	current.EffectiveAt = input.EffectiveAt.UTC()
	current.Evidence = strings.TrimSpace(input.Evidence)
	current.RelatedCode = strings.ToUpper(strings.TrimSpace(input.RelatedCode))
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.SamplingBatch{}, fmt.Errorf("update 采样批次: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "SamplingBatch", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *samplingBatchService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.SamplingBatch, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.SamplingBatch{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.SamplingBatchTransitions, current.Status, target) {
		return model.SamplingBatch{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	current.Status = target
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.SamplingBatch{}, fmt.Errorf("transition 采样批次: %w", err)
	}
	if err := s.security.Audit(ctx, actor, requestID, "transition", "SamplingBatch", id, before, target, input.Reason); err != nil {
		return model.SamplingBatch{}, fmt.Errorf("persist transition audit: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *samplingBatchService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "SamplingBatch", id, current.Status, "deleted", "soft deleted 采样批次")
}

func (s *samplingBatchService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

func validateSamplingBatchBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}
