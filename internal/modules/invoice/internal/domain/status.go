package domain

// Status represents invoice lifecycle inside the domain model.
type Status string

const (
	StatusPending    Status = "pending"
	StatusConfirming Status = "confirming"
	StatusPaid       Status = "paid"
	StatusExpired    Status = "expired"
	StatusCancelled  Status = "cancelled"
)

func (s Status) String() string { return string(s) }
