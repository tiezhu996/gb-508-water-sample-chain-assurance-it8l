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

type LabSampleService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.LabSample], error)
	Get(context.Context, uint) (model.LabSample, error)
	Create(context.Context, dto.CreateLabSample, string, string) (model.LabSample, error)
	Update(context.Context, uint, dto.UpdateLabSample, string, string) (model.LabSample, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.LabSample, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
}

type labSampleService struct {
	repository repository.LabSampleRepository
	security   SecurityService
}

func NewLabSampleService(repo repository.LabSampleRepository, security SecurityService) LabSampleService {
	return &labSampleService{repository: repo, security: security}
}

func (s *labSampleService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.LabSample], error) {
	return s.repository.List(ctx, query)
}

func (s *labSampleService) Get(ctx context.Context, id uint) (model.LabSample, error) {
	return s.repository.Get(ctx, id)
}

func (s *labSampleService) Create(ctx context.Context, input dto.CreateLabSample, actor, requestID string) (model.LabSample, error) {
	if err := validateLabSampleBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.LabSample{}, err
	}
	item := model.LabSample{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.LabSampleInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode: strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.LabSample{}, fmt.Errorf("create 实验室样本: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "LabSample", item.ID, "", item.Status, "created 实验室样本")
	return item, nil
}

func (s *labSampleService) Update(ctx context.Context, id uint, input dto.UpdateLabSample, actor, requestID string) (model.LabSample, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.LabSample{}, err
	}
	if err := validateLabSampleBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.LabSample{}, err
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
		return model.LabSample{}, fmt.Errorf("update 实验室样本: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "LabSample", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *labSampleService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.LabSample, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.LabSample{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.LabSampleTransitions, current.Status, target) {
		return model.LabSample{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	current.Status = target
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.LabSample{}, fmt.Errorf("transition 实验室样本: %w", err)
	}
	if err := s.security.Audit(ctx, actor, requestID, "transition", "LabSample", id, before, target, input.Reason); err != nil {
		return model.LabSample{}, fmt.Errorf("persist transition audit: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *labSampleService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "LabSample", id, current.Status, "deleted", "soft deleted 实验室样本")
}

func (s *labSampleService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

func validateLabSampleBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}
