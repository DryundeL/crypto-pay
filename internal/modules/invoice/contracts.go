package invoice

// Public contracts of the Invoice bounded context.
// Other modules depend only on these types/interfaces, never on invoice/internal.

// Status is the public invoice lifecycle status.
type Status string

const (
	StatusPending    Status = "pending"
	StatusConfirming Status = "confirming"
	StatusPaid       Status = "paid"
	StatusExpired    Status = "expired"
	StatusCancelled  Status = "cancelled"
)
