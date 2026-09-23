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

type AssayMethodService interface {
	List(context.Context, dto.PageQuery) (repository.Page[model.AssayMethod], error)
	Get(context.Context, uint) (model.AssayMethod, error)
	Create(context.Context, dto.CreateAssayMethod, string, string) (model.AssayMethod, error)
	Update(context.Context, uint, dto.UpdateAssayMethod, string, string) (model.AssayMethod, error)
	Transition(context.Context, uint, dto.TransitionRequest, string, string) (model.AssayMethod, error)
	Delete(context.Context, uint, string, string) error
	StatusCounts(context.Context) (map[string]int64, error)
}

type assayMethodService struct {
	repository repository.AssayMethodRepository
	security   SecurityService
}

func NewAssayMethodService(repo repository.AssayMethodRepository, security SecurityService) AssayMethodService {
	return &assayMethodService{repository: repo, security: security}
}

func (s *assayMethodService) List(ctx context.Context, query dto.PageQuery) (repository.Page[model.AssayMethod], error) {
	return s.repository.List(ctx, query)
}

func (s *assayMethodService) Get(ctx context.Context, id uint) (model.AssayMethod, error) {
	return s.repository.Get(ctx, id)
}

func (s *assayMethodService) Create(ctx context.Context, input dto.CreateAssayMethod, actor, requestID string) (model.AssayMethod, error) {
	if err := validateAssayMethodBusinessFields(input.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.AssayMethod{}, err
	}
	item := model.AssayMethod{
		BaseModel: model.BaseModel{
			Code: strings.ToUpper(strings.TrimSpace(input.Code)), Name: strings.TrimSpace(input.Name),
			Status: model.AssayMethodInitialStatus, Version: 1, Description: strings.TrimSpace(input.Description),
		},
		Facility: strings.TrimSpace(input.Facility), Owner: strings.TrimSpace(input.Owner),
		Category: strings.TrimSpace(input.Category), RiskLevel: input.RiskLevel,
		MetricValue: input.MetricValue, MetricUnit: strings.TrimSpace(input.MetricUnit),
		EffectiveAt: input.EffectiveAt.UTC(), Evidence: strings.TrimSpace(input.Evidence),
		RelatedCode: strings.ToUpper(strings.TrimSpace(input.RelatedCode)),
	}
	if err := s.repository.Create(ctx, &item); err != nil {
		return model.AssayMethod{}, fmt.Errorf("create 检测方法: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "create", "AssayMethod", item.ID, "", item.Status, "created 检测方法")
	return item, nil
}

func (s *assayMethodService) Update(ctx context.Context, id uint, input dto.UpdateAssayMethod, actor, requestID string) (model.AssayMethod, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.AssayMethod{}, err
	}
	if err := validateAssayMethodBusinessFields(current.Code, input.Name, input.Facility, input.Owner); err != nil {
		return model.AssayMethod{}, err
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
		return model.AssayMethod{}, fmt.Errorf("update 检测方法: %w", err)
	}
	_ = s.security.Audit(ctx, actor, requestID, "update", "AssayMethod", id, current.Status, current.Status, "updated business fields")
	return s.repository.Get(ctx, id)
}

func (s *assayMethodService) Transition(ctx context.Context, id uint, input dto.TransitionRequest, actor, requestID string) (model.AssayMethod, error) {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return model.AssayMethod{}, err
	}
	target := strings.TrimSpace(input.Status)
	if !constants.CanTransition(constants.AssayMethodTransitions, current.Status, target) {
		return model.AssayMethod{}, fmt.Errorf("%w: %s -> %s", ErrInvalidTransition, current.Status, target)
	}
	before := current.Status
	current.Status = target
	current.Version = input.ExpectedVersion + 1
	current.UpdatedAt = time.Now().UTC()
	if err := s.repository.Update(ctx, id, input.ExpectedVersion, &current); err != nil {
		return model.AssayMethod{}, fmt.Errorf("transition 检测方法: %w", err)
	}
	if err := s.security.Audit(ctx, actor, requestID, "transition", "AssayMethod", id, before, target, input.Reason); err != nil {
		return model.AssayMethod{}, fmt.Errorf("persist transition audit: %w", err)
	}
	return s.repository.Get(ctx, id)
}

func (s *assayMethodService) Delete(ctx context.Context, id uint, actor, requestID string) error {
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return s.security.Audit(ctx, actor, requestID, "delete", "AssayMethod", id, current.Status, "deleted", "soft deleted 检测方法")
}

func (s *assayMethodService) StatusCounts(ctx context.Context) (map[string]int64, error) {
	return s.repository.CountByStatus(ctx)
}

func validateAssayMethodBusinessFields(code, name, facility, owner string) error {
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(facility) == "" || strings.TrimSpace(owner) == "" {
		return ErrInvalidInput
	}
	return nil
}
