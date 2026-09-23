package service

import (
	"errors"
	"testing"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
)

func TestValidateResultReviewTransitionRejectsSameSigner(t *testing.T) {
	current := model.ResultReview{
		BaseModel:         model.BaseModel{Status: "peer_review"},
		ReviewRequestedBy: "reviewer-a",
	}

	err := validateResultReviewTransition(current, "signed", "reviewer-a")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected same-person signing to fail business validation, got %v", err)
	}
	if err := validateResultReviewTransition(current, "signed", "reviewer-b"); err != nil {
		t.Fatalf("expected a different reviewer to sign: %v", err)
	}
}
