package domain

// Status is the withdrawal lifecycle inside the domain model.
// v1 runtime: requested | completed | rejected.
type Status string

const (
	StatusRequested Status = "requested"
	StatusCompleted Status = "completed"
	StatusRejected  Status = "rejected"
)

func (s Status) String() string { return string(s) }

func (s Status) Valid() bool {
	switch s {
	case StatusRequested, StatusCompleted, StatusRejected:
		return true
	default:
		return false
	}
}
