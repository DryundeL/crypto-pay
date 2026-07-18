package domain

type Status string

const (
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

func (s Status) String() string { return string(s) }

type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
)

func (s APIKeyStatus) String() string { return string(s) }
