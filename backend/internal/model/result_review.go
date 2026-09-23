package model

import "time"

// ResultReview models 结果复核 as an independently versioned aggregate. The fields
// cover ownership, operational context, evidence and measured risk so later
// changes naturally span persistence, service and UI layers.
type ResultReview struct {
	BaseModel
	Facility          string    `json:"facility" gorm:"size:120;index"`
	Owner             string    `json:"owner" gorm:"size:120;index"`
	Category          string    `json:"category" gorm:"size:80;index"`
	RiskLevel         string    `json:"riskLevel" gorm:"size:32;index"`
	MetricValue       float64   `json:"metricValue"`
	MetricUnit        string    `json:"metricUnit" gorm:"size:24"`
	EffectiveAt       time.Time `json:"effectiveAt"`
	Evidence          string    `json:"evidence" gorm:"size:2000"`
	RelatedCode       string    `json:"relatedCode" gorm:"size:64;index"`
	ReviewRequestedBy string    `json:"reviewRequestedBy" gorm:"size:80;index"`
	PeerReviewedBy    string    `json:"peerReviewedBy" gorm:"size:80;index"`
	SignedBy          string    `json:"signedBy" gorm:"size:80;index"`
}

func (item *ResultReview) GetBase() *BaseModel { return &item.BaseModel }

func (item ResultReview) TableName() string { return "result_reviews" }

var ResultReviewInitialStatus = "draft"
