package domain

type WithdrawalID string

func (id WithdrawalID) String() string { return string(id) }
func (id WithdrawalID) IsZero() bool   { return id == "" }
