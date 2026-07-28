package command

type RecordObserved struct {
	MerchantID    string // auth merchant; must own matched invoice
	Network       string
	TxHash        string
	ToAddress     string
	Amount        string
	Currency      string
	Confirmations int
}
