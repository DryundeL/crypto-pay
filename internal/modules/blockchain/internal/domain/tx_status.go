package domain

// TxStatus is the lifecycle of a watched chain transaction.
type TxStatus string

const (
	TxStatusObserved  TxStatus = "observed"
	TxStatusConfirmed TxStatus = "confirmed"
)

func (s TxStatus) String() string { return string(s) }
