package domain

import "errors"

var (
	ErrInvalidMerchant   = errors.New("invalid merchant")
	ErrMerchantNotFound  = errors.New("merchant not found")
	ErrMerchantSuspended = errors.New("merchant is suspended")
	ErrInvalidAPIKey     = errors.New("invalid api key")
	ErrAPIKeyNotFound    = errors.New("api key not found")
	ErrAPIKeyRevoked     = errors.New("api key already revoked")
)
