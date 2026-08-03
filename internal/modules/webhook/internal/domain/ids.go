package domain

type DeliveryID string

func (id DeliveryID) String() string { return string(id) }
func (id DeliveryID) IsZero() bool   { return id == "" }
