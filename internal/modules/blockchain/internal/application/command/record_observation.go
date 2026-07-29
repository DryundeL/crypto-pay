package command

type RecordObservation struct {
	MerchantID    string
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
}
