package domain

import "errors"

var (
	ErrInvalidDelivery   = errors.New("invalid webhook delivery")
	ErrDeliveryNotFound  = errors.New("webhook delivery not found")
	ErrInvalidTransition = errors.New("invalid webhook delivery status transition")
)
