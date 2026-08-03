package ports

import "context"

type DeliverResult struct {
	StatusCode int
	Retryable  bool
	Error      string
}

// HTTPDeliverer POSTs a signed webhook payload to the merchant URL.
type HTTPDeliverer interface {
	Deliver(ctx context.Context, deliveryID, eventName, url string, payload []byte) (DeliverResult, error)
}
