package command

type ConfirmPayment struct {
	MerchantID string
	Network    string
	TxHash     string
}
