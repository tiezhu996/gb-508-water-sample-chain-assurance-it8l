package constants

// Shared status values are mirrored in frontend/src/types/status.ts. Keeping
// the lists explicit makes state-machine drift visible during code review.

type SampleState string

const (
	SampleStateReceived SampleState = "received"
	SampleStateAccepted SampleState = "accepted"
	SampleStateTesting  SampleState = "testing"
	SampleStateHold     SampleState = "hold"
	SampleStateDisposed SampleState = "disposed"
)

var AllSampleState = []string{"received", "accepted", "testing", "hold", "disposed"}

type ReviewState string

const (
	ReviewStateDraft      ReviewState = "draft"
	ReviewStatePeerReview ReviewState = "peer_review"
	ReviewStateSigned     ReviewState = "signed"
	ReviewStateRejected   ReviewState = "rejected"
)

var AllReviewState = []string{"draft", "peer_review", "signed", "rejected"}

var SamplingBatchTransitions = map[string]map[string]bool{
	"planned":    {"collecting": true, "received": true},
	"collecting": {"received": true, "closed": true, "planned": true},
	"received":   {"closed": true, "collecting": true},
	"closed":     {"received": true},
}

var LabSampleTransitions = map[string]map[string]bool{
	"received": {"accepted": true, "testing": true},
	"accepted": {"testing": true, "hold": true, "received": true},
	"testing":  {"hold": true, "disposed": true, "accepted": true},
	"hold":     {"disposed": true, "testing": true},
	"disposed": {"hold": true},
}

var AssayMethodTransitions = map[string]map[string]bool{
	"draft":     {"validated": true, "active": true},
	"validated": {"active": true, "retired": true, "draft": true},
	"active":    {"retired": true, "validated": true},
	"retired":   {"active": true},
}

var ResultReviewTransitions = map[string]map[string]bool{
	"draft":       {"peer_review": true},
	"peer_review": {"signed": true, "rejected": true, "draft": true},
	"signed":      {"rejected": true},
	"rejected":    {"draft": true, "peer_review": true},
}

func CanTransition(graph map[string]map[string]bool, from, to string) bool {
	targets, exists := graph[from]
	return exists && targets[to]
}
