package domain

// Invoice is the aggregate root of the Invoice bounded context.
type Invoice struct {
	id         InvoiceID
	merchantID string
	status     Status
	amount     Money
}

func (i *Invoice) ID() InvoiceID   { return i.id }
func (i *Invoice) Status() Status  { return i.status }
func (i *Invoice) Amount() Money   { return i.amount }
func (i *Invoice) MerchantID() string { return i.merchantID }
