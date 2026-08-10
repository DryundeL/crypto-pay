package ports

// ConfirmationPolicy returns the minimum confirmations required per network.
type ConfirmationPolicy interface {
	Required(network string) (int, error)
}
