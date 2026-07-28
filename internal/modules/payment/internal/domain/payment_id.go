package domain

// PaymentID is a typed identity for Payment aggregate.
type PaymentID string

func (id PaymentID) String() string { return string(id) }

func (id PaymentID) IsZero() bool { return id == "" }
