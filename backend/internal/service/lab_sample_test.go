package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/config"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/constants"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var fixtureCounter atomic.Int64

func TestValidateCustodyHandoverRules(t *testing.T) {
	current := model.LabSample{
		BaseModel: model.BaseModel{ID: 1, Status: string(constants.SampleStateAccepted), Version: 3},
		Custodian: "operator",
	}
	enabledOperator := model.User{Username: "reviewer", Role: model.RoleReviewer, Active: true}

	cases := []struct {
		name    string
		sample  model.LabSample
		actor   string
		target  string
		account model.User
		wantErr error
	}{
		{name: "happy path", sample: current, actor: "operator", target: "reviewer", account: enabledOperator},
		{name: "disposed sample", sample: custodianSample(string(constants.SampleStateDisposed), "operator"), actor: "operator", target: "reviewer", account: enabledOperator, wantErr: ErrCustodyDisposed},
		{name: "not current custodian", sample: current, actor: "someone-else", target: "reviewer", account: enabledOperator, wantErr: ErrCustodyNotCustodian},
		{name: "same custodian and target", sample: current, actor: "operator", target: "operator", account: enabledOperator, wantErr: ErrCustodySameTarget},
		{name: "inactive target account", sample: current, actor: "operator", target: "retired-ops", account: model.User{Username: "retired-ops", Role: model.RoleOperator, Active: false}, wantErr: ErrCustodyAccount},
		{name: "target role below operator", sample: current, actor: "operator", target: "viewer", account: model.User{Username: "viewer", Role: model.RoleViewer, Active: true}, wantErr: ErrCustodyTargetRole},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCustodyHandover(tc.sample, tc.actor, tc.target, tc.account)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("validateCustodyHandover() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func custodianSample(status, custodian string) model.LabSample {
	return model.LabSample{
		BaseModel: model.BaseModel{ID: 1, Status: string(status), Version: 1},
		Custodian: custodian,
	}
}

func newCustodyFixture(t *testing.T) (context.Context, LabSampleService, SecurityService, *gorm.DB) {
	t.Helper()
	// Unique shared-cache DSN gives every test case an isolated in-memory DB
	// while keeping all pooled connections pointed at the same database.
	dsn := fmt.Sprintf("file:custody-%s-%d?mode=memory&cache=shared", t.Name(), fixtureCounter.Add(1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.AuditLog{}, &model.LabSample{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	securityRepository := repository.NewSecurityRepository(db)
	security := NewSecurityService(securityRepository, config.Config{})
	labRepository := repository.NewLabSampleRepository(db)
	service := NewLabSampleService(labRepository, security)
	ctx := context.Background()
	users := []model.User{
		{Username: "operator", DisplayName: "现场操作员", PasswordHash: "x", Role: model.RoleOperator, Active: true},
		{Username: "reviewer", DisplayName: "质量复核员", PasswordHash: "x", Role: model.RoleReviewer, Active: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	sample := model.LabSample{
		BaseModel: model.BaseModel{Code: "LS-T1", Name: "并发交接样本", Status: string(constants.SampleStateTesting), Version: 1},
		Facility:  "实验室A", Owner: "运行一组", Category: "常规", RiskLevel: "medium",
		EffectiveAt: time.Now().UTC(), Custodian: "operator",
	}
	if err := labRepository.Create(ctx, &sample); err != nil {
		t.Fatalf("seed sample: %v", err)
	}
	return ctx, service, security, db
}

func TestHandoverSucceedsAndSyncsFields(t *testing.T) {
	ctx, service, _, _ := newCustodyFixture(t)
	updated, err := service.Handover(ctx, 1, dto.HandoverLabSample{TargetUsername: "reviewer", ExpectedVersion: 1, Reason: "转复核"}, "operator", "req-1")
	if err != nil {
		t.Fatalf("handover: %v", err)
	}
	if updated.Custodian != "reviewer" {
		t.Fatalf("custodian = %q, want reviewer", updated.Custodian)
	}
	if updated.Version != 2 {
		t.Fatalf("version = %d, want 2", updated.Version)
	}
	if updated.CustodyTransferredAt == nil {
		t.Fatal("custodyTransferredAt should be recorded")
	}
	if updated.Status != string(constants.SampleStateTesting) {
		t.Fatalf("status changed to %q, handover must keep status", updated.Status)
	}
	reread, err := service.Get(ctx, 1)
	if err != nil {
		t.Fatalf("reread: %v", err)
	}
	if reread.Custodian != "reviewer" || reread.Version != 2 || reread.CustodyTransferredAt == nil {
		t.Fatalf("handover not persisted: %+v", reread)
	}
}

func TestHandoverRejectedLeavesRecordUntouched(t *testing.T) {
	ctx, service, _, _ := newCustodyFixture(t)
	_, err := service.Handover(ctx, 1, dto.HandoverLabSample{TargetUsername: "operator", ExpectedVersion: 1}, "operator", "req-2")
	if !errors.Is(err, ErrCustodySameTarget) {
		t.Fatalf("want same-target rejection, got %v", err)
	}
	current, _ := service.Get(ctx, 1)
	if current.Custodian != "operator" || current.Version != 1 || current.CustodyTransferredAt != nil {
		t.Fatalf("rejected handover mutated the record: %+v", current)
	}
}

// TestConcurrentHandoverSucceedsOnce fires identical requests at the same
// version. Exactly one must win. Losers either miss the version guard
// (ErrVersionConflict when they snapshotted v1 together with the winner) or, if
// they only read the record after the winner committed, are rejected as a
// non-custodian against the now-current data. In both cases the record stays
// consistent and no second handover occurs.
func TestConcurrentHandoverSucceedsOnce(t *testing.T) {
	ctx, service, _, _ := newCustodyFixture(t)
	const workers = 8
	var wg sync.WaitGroup
	results := make(chan error, workers)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := service.Handover(ctx, 1, dto.HandoverLabSample{TargetUsername: "reviewer", ExpectedVersion: 1}, "operator", "req-concurrent")
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, repository.ErrVersionConflict), errors.Is(err, ErrCustodyNotCustodian):
			// expected losing outcome, see comment above
		default:
			t.Fatalf("unexpected concurrent handover error: %v", err)
		}
	}
	if successes != 1 {
		t.Fatalf("successes = %d, want exactly 1", successes)
	}
	final, _ := service.Get(ctx, 1)
	if final.Custodian != "reviewer" || final.Version != 2 {
		t.Fatalf("final record = custodian %q version %d, want reviewer/2", final.Custodian, final.Version)
	}
}
