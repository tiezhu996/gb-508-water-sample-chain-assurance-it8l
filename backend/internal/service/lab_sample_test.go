package service

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/config"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/dto"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var handoverDBSuffix int64
var sampleCodeSuffix int64

func newLabSampleServiceWithDB(t *testing.T) (LabSampleService, *gorm.DB) {
	t.Helper()
	dsn := fmt.Sprintf("file:handover-%d?mode=memory&cache=shared", atomic.AddInt64(&handoverDBSuffix, 1))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.AuditLog{}, &model.LabSample{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Migrator().DropTable("audit_logs", "lab_samples", "users")
	})
	securityRepo := repository.NewSecurityRepository(db)
	security := NewSecurityService(securityRepo, config.Config{})
	return NewLabSampleService(repository.NewLabSampleRepository(db), security), db
}

func seedHandoverUsers(t *testing.T, db *gorm.DB) {
	t.Helper()
	users := []model.User{
		{Username: "operator", DisplayName: "操作员", PasswordHash: "x", Role: model.RoleOperator, Active: true},
		{Username: "reviewer", DisplayName: "复核员", PasswordHash: "x", Role: model.RoleReviewer, Active: true},
		{Username: "viewer", DisplayName: "只读", PasswordHash: "x", Role: model.RoleViewer, Active: true},
		{Username: "disabled", DisplayName: "停用", PasswordHash: "x", Role: model.RoleOperator, Active: true},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}
	// The Active column declares DEFAULT true, so force the zero value with an
	// explicit column update to emulate a deactivated account.
	if err := db.Model(&model.User{}).Where("username = ?", "disabled").Update("active", false).Error; err != nil {
		t.Fatalf("deactivate test user: %v", err)
	}
}

func seedSample(t *testing.T, svc LabSampleService, status, custodian string) model.LabSample {
	t.Helper()
	ctx := context.Background()
	code := fmt.Sprintf("LS-%d", atomic.AddInt64(&sampleCodeSuffix, 1))
	item := model.LabSample{
		BaseModel: model.BaseModel{Code: code, Name: "交接测试样本", Status: status, Version: 1},
		Facility:  "一号实验室", Owner: "运行一组", Category: "常规", RiskLevel: "low",
		EffectiveAt: time.Now().UTC(), Custodian: custodian,
	}
	if err := svc.(*labSampleService).repository.Create(ctx, &item); err != nil {
		t.Fatalf("seed sample: %v", err)
	}
	return item
}

func TestHandoverSucceedsAndSyncsCustodianTimeVersionAndAudit(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)
	sample := seedSample(t, svc, "testing", "operator")

	result, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
		TargetUsername: "reviewer", ExpectedVersion: 1, Remark: "移交复核",
	}, "operator", "req-1")
	if err != nil {
		t.Fatalf("expected handover to succeed: %v", err)
	}
	if result.Custodian != "reviewer" {
		t.Fatalf("custodian = %q, want reviewer", result.Custodian)
	}
	if result.Version != 2 {
		t.Fatalf("version = %d, want 2", result.Version)
	}
	if result.HandoverAt == nil {
		t.Fatal("expected handoverAt to be recorded")
	}
	if result.Status != "testing" {
		t.Fatalf("status must remain unchanged, got %q", result.Status)
	}
	var audit model.AuditLog
	if err := db.Where("action = ? AND entity_id = ?", "handover", sample.ID).First(&audit).Error; err != nil {
		t.Fatalf("expected handover audit row: %v", err)
	}
	if audit.BeforeState != "operator" || audit.AfterState != "reviewer" {
		t.Fatalf("audit chain = %q -> %q", audit.BeforeState, audit.AfterState)
	}
	if audit.RequestID != "req-1" || audit.Actor != "operator" {
		t.Fatalf("audit metadata wrong: %+v", audit)
	}
}

func TestHandoverRejectionsLeaveOriginalRecord(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)

	cases := []struct {
		name   string
		actor  string
		target string
		wrap   error
	}{
		{"non custodian", "reviewer", "reviewer", ErrCustodianMismatch},
		{"target is self", "operator", "operator", ErrInvalidInput},
		{"target missing", "operator", "ghost", ErrInvalidInput},
		{"target inactive", "operator", "disabled", ErrInvalidInput},
		{"target role too low", "operator", "viewer", ErrInvalidInput},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Every subtest gets a pristine sample so a rejected case can never
			// influence the guard evaluation of a later case.
			sample := seedSample(t, svc, "accepted", "operator")
			_, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
				TargetUsername: tc.target, ExpectedVersion: 1,
			}, tc.actor, "req-blocked")
			if !errors.Is(err, tc.wrap) {
				t.Fatalf("expected %v, got %v", tc.wrap, err)
			}
			current, getErr := svc.Get(context.Background(), sample.ID)
			if getErr != nil {
				t.Fatalf("reload sample: %v", getErr)
			}
			if current.Custodian != "operator" || current.Version != 1 || current.Status != "accepted" || current.HandoverAt != nil {
				t.Fatalf("original record changed after rejection: %+v", current)
			}
		})
	}

	var audits int64
	if err := db.Model(&model.AuditLog{}).Where("action = ?", "handover").Count(&audits).Error; err != nil {
		t.Fatalf("count audits: %v", err)
	}
	if audits != 0 {
		t.Fatalf("rejected handovers must not write audits, found %d", audits)
	}
}

func TestHandoverRejectsDisposedSample(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)
	sample := seedSample(t, svc, "disposed", "operator")

	_, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
		TargetUsername: "reviewer", ExpectedVersion: 1,
	}, "operator", "req-disposed")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected disposed sample to be rejected, got %v", err)
	}
	current, _ := svc.Get(context.Background(), sample.ID)
	if current.Custodian != "operator" || current.HandoverAt != nil {
		t.Fatalf("disposed sample record changed: %+v", current)
	}
}

func TestHandoverConcurrentVersionWinsOnce(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)
	sample := seedSample(t, svc, "received", "operator")

	// Two requests carrying the same expected version race to the database:
	// the optimistic-lock predicate allows exactly one of them to update.
	start := make(chan struct{})
	errs := make(chan error, 2)
	go func() {
		<-start
		_, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
			TargetUsername: "reviewer", ExpectedVersion: sample.Version,
		}, "operator", "req-racer-a")
		errs <- err
	}()
	go func() {
		<-start
		_, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
			TargetUsername: "reviewer", ExpectedVersion: sample.Version,
		}, "operator", "req-racer-b")
		errs <- err
	}()
	close(start)

	conflicts, successes, rejections := 0, 0, 0
	for i := 0; i < 2; i++ {
		switch err := <-errs; {
		case err == nil:
			successes++
		case errors.Is(err, repository.ErrVersionConflict):
			conflicts++
			rejections++
		default:
			// A serialized store may let the loser observe the already-committed
			// custody change; that rejection still guarantees a single winner.
			rejections++
		}
	}
	if successes != 1 || rejections != 1 {
		t.Fatalf("expected 1 success and 1 rejection, got %d success / %d rejection (%d version conflicts)", successes, rejections, conflicts)
	}

	current, _ := svc.Get(context.Background(), sample.ID)
	if current.Custodian != "reviewer" || current.Version != 2 {
		t.Fatalf("unexpected record after concurrent handover: %+v", current)
	}
	var audits int64
	if err := db.Model(&model.AuditLog{}).Where("action = ?", "handover").Count(&audits).Error; err != nil {
		t.Fatalf("count audits: %v", err)
	}
	if audits != 1 {
		t.Fatalf("expected exactly one handover audit, got %d", audits)
	}
}

func TestHandoverSequentialStaleRequestIsRejected(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)
	sample := seedSample(t, svc, "testing", "operator")

	if _, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
		TargetUsername: "reviewer", ExpectedVersion: 1,
	}, "operator", "req-first"); err != nil {
		t.Fatalf("first handover failed: %v", err)
	}
	// A stale request submitted by the old custodian must be refused without a
	// second write or audit row.
	_, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
		TargetUsername: "operator", ExpectedVersion: 1,
	}, "operator", "req-stale")
	if err == nil {
		t.Fatal("expected stale sequential handover to be rejected")
	}
	current, _ := svc.Get(context.Background(), sample.ID)
	if current.Custodian != "reviewer" || current.Version != 2 {
		t.Fatalf("stale request altered the record: %+v", current)
	}
	var audits int64
	_ = db.Model(&model.AuditLog{}).Where("action = ?", "handover").Count(&audits)
	if audits != 1 {
		t.Fatalf("expected exactly one handover audit, got %d", audits)
	}
}

func TestHandoverRoundTripsAfterRefresh(t *testing.T) {
	svc, db := newLabSampleServiceWithDB(t)
	seedHandoverUsers(t, db)
	sample := seedSample(t, svc, "hold", "operator")

	if _, err := svc.Handover(context.Background(), sample.ID, dto.HandoverLabSample{
		TargetUsername: "reviewer", ExpectedVersion: 1,
	}, "operator", "req-roundtrip"); err != nil {
		t.Fatalf("handover failed: %v", err)
	}
	reloaded, err := svc.Get(context.Background(), sample.ID)
	if err != nil {
		t.Fatalf("reload after handover: %v", err)
	}
	if reloaded.Custodian != "reviewer" || reloaded.HandoverAt == nil || reloaded.Version != 2 {
		t.Fatalf("handover state did not persist: %+v", reloaded)
	}
}
