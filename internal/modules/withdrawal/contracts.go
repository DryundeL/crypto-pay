package withdrawal

// Public contracts of the Withdrawal bounded context.
type Status string

const (
	StatusRequested Status = "requested"
	StatusApproved  Status = "approved"
	StatusBroadcast Status = "broadcast"
	StatusCompleted Status = "completed"
	StatusRejected  Status = "rejected"
)
