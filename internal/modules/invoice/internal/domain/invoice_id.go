package domain

// InvoiceID is a typed identity for Invoice aggregate.
type InvoiceID string

func (id InvoiceID) String() string { return string(id) }

func (id InvoiceID) IsZero() bool { return id == "" }
