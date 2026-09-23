package model

import "time"

// LabSample models 实验室样本 as an independently versioned aggregate. The fields
// cover ownership, operational context, evidence and measured risk so later
// changes naturally span persistence, service and UI layers.
//
// Custodian / CustodyTransferredAt track the current physical keeper of the
// sample. Custody changes do not alter the operational Status: they are a
// separate chain-of-custody move with their own audit entry and optimistic
// version, so a handed-over sample keeps its received/accepted/testing flow.
type LabSample struct {
	BaseModel
	Facility             string     `json:"facility" gorm:"size:120;index"`
	Owner                string     `json:"owner" gorm:"size:120;index"`
	Category             string     `json:"category" gorm:"size:80;index"`
	RiskLevel            string     `json:"riskLevel" gorm:"size:32;index"`
	MetricValue          float64    `json:"metricValue"`
	MetricUnit           string     `json:"metricUnit" gorm:"size:24"`
	EffectiveAt          time.Time  `json:"effectiveAt"`
	Evidence             string     `json:"evidence" gorm:"size:2000"`
	RelatedCode          string     `json:"relatedCode" gorm:"size:64;index"`
	Custodian            string     `json:"custodian" gorm:"size:80;index"`
	CustodyTransferredAt *time.Time `json:"custodyTransferredAt"`
}

func (item *LabSample) GetBase() *BaseModel { return &item.BaseModel }

func (item LabSample) TableName() string { return "lab_samples" }

var LabSampleInitialStatus = "received"
