package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/blueship581/water-sample-chain-assurance/backend/internal/config"
	"github.com/blueship581/water-sample-chain-assurance/backend/internal/model"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(ctx context.Context, cfg config.Config, log *slog.Logger) (*gorm.DB, *redis.Client, error) {
	var dialector gorm.Dialector
	switch cfg.DatabaseDriver {
	case "postgres":
		dialector = postgres.Open(cfg.DatabaseDSN)
	case "mysql":
		dialector = mysql.Open(cfg.DatabaseDSN)
	case "sqlite":
		dialector = sqlite.Open(cfg.DatabaseDSN)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q", cfg.DatabaseDriver)
	}
	logLevel := logger.Warn
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}
	var db *gorm.DB
	var err error
	for attempt := 1; attempt <= 20; attempt++ {
		db, err = gorm.Open(dialector, &gorm.Config{Logger: logger.Default.LogMode(logLevel)})
		if err == nil {
			sqlDB, dbErr := db.DB()
			if dbErr == nil && sqlDB.PingContext(ctx) == nil {
				break
			}
			if dbErr != nil {
				err = dbErr
			} else {
				err = sqlDB.PingContext(ctx)
			}
		}
		log.Warn("database not ready", "attempt", attempt, "error", err)
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	if err != nil {
		return nil, nil, fmt.Errorf("connect database: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, nil, err
	}
	if err := Seed(ctx, db); err != nil {
		return nil, nil, err
	}
	var redisClient *redis.Client
	if cfg.RedisAddr != "" {
		redisClient = redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword})
		if err := redisClient.Ping(ctx).Err(); err != nil {
			return nil, nil, fmt.Errorf("connect redis: %w", err)
		}
	}
	return db, redisClient, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{}, &model.AuditLog{},
		&model.SamplingBatch{},
		&model.LabSample{},
		&model.AssayMethod{},
		&model.ResultReview{},
	)
}

func Seed(ctx context.Context, db *gorm.DB) error {
	var users int64
	if err := db.WithContext(ctx).Model(&model.User{}).Count(&users).Error; err != nil {
		return err
	}
	if users == 0 {
		password, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		seedUsers := []model.User{
			{Username: "admin", DisplayName: "系统管理员", PasswordHash: string(password), Role: model.RoleAdmin, Active: true},
			{Username: "reviewer", DisplayName: "质量复核员", PasswordHash: string(password), Role: model.RoleReviewer, Active: true},
			{Username: "operator", DisplayName: "现场操作员", PasswordHash: string(password), Role: model.RoleOperator, Active: true},
		}
		if err := db.WithContext(ctx).Create(&seedUsers).Error; err != nil {
			return err
		}
	}

	if err := seedSamplingBatch(ctx, db); err != nil {
		return err
	}

	if err := seedLabSample(ctx, db); err != nil {
		return err
	}

	if err := seedAssayMethod(ctx, db); err != nil {
		return err
	}

	if err := seedResultReview(ctx, db); err != nil {
		return err
	}

	return nil
}

func seedSamplingBatch(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.SamplingBatch{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.SamplingBatch{

		{BaseModel: model.BaseModel{Code: "SB-001", Name: "采样批次示例一", Status: "planned", Version: 1,
			Description: "用于启动验证和主要流程演示的采样批次记录"}, Facility: "水质检测样本链路审核区域1", Owner: "运行一组",
			Category: "常规", RiskLevel: "low", MetricValue: 12.5, MetricUnit: "unit",
			EffectiveAt: now.Add(0 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-01"},

		{BaseModel: model.BaseModel{Code: "SB-002", Name: "采样批次示例二", Status: "collecting", Version: 1,
			Description: "用于启动验证和主要流程演示的采样批次记录"}, Facility: "水质检测样本链路审核区域2", Owner: "质量复核组",
			Category: "重点", RiskLevel: "medium", MetricValue: 25.0, MetricUnit: "%",
			EffectiveAt: now.Add(3 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-02"},

		{BaseModel: model.BaseModel{Code: "SB-003", Name: "采样批次示例三", Status: "received", Version: 1,
			Description: "用于启动验证和主要流程演示的采样批次记录"}, Facility: "水质检测样本链路审核区域3", Owner: "安全主管组",
			Category: "复核", RiskLevel: "high", MetricValue: 37.5, MetricUnit: "score",
			EffectiveAt: now.Add(6 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-03"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedLabSample(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.LabSample{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.LabSample{

		{BaseModel: model.BaseModel{Code: "LS-001", Name: "实验室样本示例一", Status: "received", Version: 1,
			Description: "用于启动验证和主要流程演示的实验室样本记录"}, Facility: "水质检测样本链路审核区域1", Owner: "运行一组",
			Category: "常规", RiskLevel: "low", MetricValue: 12.5, MetricUnit: "unit",
			EffectiveAt: now.Add(0 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-01", Custodian: "operator"},

		{BaseModel: model.BaseModel{Code: "LS-002", Name: "实验室样本示例二", Status: "accepted", Version: 1,
			Description: "用于启动验证和主要流程演示的实验室样本记录"}, Facility: "水质检测样本链路审核区域2", Owner: "质量复核组",
			Category: "重点", RiskLevel: "medium", MetricValue: 25.0, MetricUnit: "%",
			EffectiveAt: now.Add(3 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-02", Custodian: "operator"},

		{BaseModel: model.BaseModel{Code: "LS-003", Name: "实验室样本示例三", Status: "testing", Version: 1,
			Description: "用于启动验证和主要流程演示的实验室样本记录"}, Facility: "水质检测样本链路审核区域3", Owner: "安全主管组",
			Category: "复核", RiskLevel: "high", MetricValue: 37.5, MetricUnit: "score",
			EffectiveAt: now.Add(6 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-03", Custodian: "operator"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedAssayMethod(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.AssayMethod{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.AssayMethod{

		{BaseModel: model.BaseModel{Code: "AM-001", Name: "检测方法示例一", Status: "draft", Version: 1,
			Description: "用于启动验证和主要流程演示的检测方法记录"}, Facility: "水质检测样本链路审核区域1", Owner: "运行一组",
			Category: "常规", RiskLevel: "low", MetricValue: 12.5, MetricUnit: "unit",
			EffectiveAt: now.Add(0 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-01"},

		{BaseModel: model.BaseModel{Code: "AM-002", Name: "检测方法示例二", Status: "validated", Version: 1,
			Description: "用于启动验证和主要流程演示的检测方法记录"}, Facility: "水质检测样本链路审核区域2", Owner: "质量复核组",
			Category: "重点", RiskLevel: "medium", MetricValue: 25.0, MetricUnit: "%",
			EffectiveAt: now.Add(3 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-02"},

		{BaseModel: model.BaseModel{Code: "AM-003", Name: "检测方法示例三", Status: "active", Version: 1,
			Description: "用于启动验证和主要流程演示的检测方法记录"}, Facility: "水质检测样本链路审核区域3", Owner: "安全主管组",
			Category: "复核", RiskLevel: "high", MetricValue: 37.5, MetricUnit: "score",
			EffectiveAt: now.Add(6 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "REL-508-03"},
	}
	return db.WithContext(ctx).Create(&items).Error
}

func seedResultReview(ctx context.Context, db *gorm.DB) error {
	var count int64
	if err := db.WithContext(ctx).Model(&model.ResultReview{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now().UTC()
	items := []model.ResultReview{

		{BaseModel: model.BaseModel{Code: "RR-001", Name: "结果复核示例一", Status: "draft", Version: 1,
			Description: "用于启动验证和主要流程演示的结果复核记录"}, Facility: "水质检测样本链路审核区域1", Owner: "运行一组",
			Category: "常规", RiskLevel: "low", MetricValue: 12.5, MetricUnit: "unit",
			EffectiveAt: now.Add(0 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "AM-001"},

		{BaseModel: model.BaseModel{Code: "RR-002", Name: "结果复核示例二", Status: "peer_review", Version: 1,
			Description: "用于启动验证和主要流程演示的结果复核记录"}, Facility: "水质检测样本链路审核区域2", Owner: "质量复核组",
			Category: "重点", RiskLevel: "medium", MetricValue: 25.0, MetricUnit: "%",
			EffectiveAt: now.Add(3 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "AM-002", ReviewRequestedBy: "operator"},

		{BaseModel: model.BaseModel{Code: "RR-003", Name: "结果复核示例三", Status: "signed", Version: 1,
			Description: "用于启动验证和主要流程演示的结果复核记录"}, Facility: "水质检测样本链路审核区域3", Owner: "安全主管组",
			Category: "复核", RiskLevel: "high", MetricValue: 37.5, MetricUnit: "score",
			EffectiveAt: now.Add(6 * time.Hour), Evidence: "已完成基础证据核对", RelatedCode: "AM-003", ReviewRequestedBy: "operator", PeerReviewedBy: "reviewer", SignedBy: "reviewer"},
	}
	return db.WithContext(ctx).Create(&items).Error
}
