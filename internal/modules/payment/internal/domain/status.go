package domain

// Status represents payment lifecycle inside the domain model.
type Status string

const (
	StatusDetected   Status = "detected"
	StatusConfirming Status = "confirming"
	StatusConfirmed  Status = "confirmed"
	StatusFailed     Status = "failed"
)

func (s Status) String() string { return string(s) }
