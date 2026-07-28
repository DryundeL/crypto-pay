package command

type CreateInvoice struct {
	MerchantID       string
	Amount           string
	Currency         string
	Network          string
	ExpiresInSeconds int // 0 = default TTL
}
