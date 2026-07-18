package payment

// Public contracts of the Payment bounded context.
type Status string

const (
	StatusDetected   Status = "detected"
	StatusConfirming Status = "confirming"
	StatusConfirmed  Status = "confirmed"
	StatusFailed     Status = "failed"
)
