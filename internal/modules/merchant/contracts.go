package merchant

// Public contracts of the Merchant bounded context.

type ID string

func (id ID) String() string { return string(id) }

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)
