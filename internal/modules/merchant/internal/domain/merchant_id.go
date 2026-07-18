package domain

type MerchantID string

func (id MerchantID) String() string { return string(id) }

func (id MerchantID) IsZero() bool { return id == "" }
