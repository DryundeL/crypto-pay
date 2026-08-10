package payment

import (
	"fmt"
	"strings"
)

// ConfirmationPolicy maps network → required confirmations for Confirm.
type ConfirmationPolicy struct {
	required map[string]int
}

// NewConfirmationPolicy builds a policy from network thresholds.
// Unknown / empty networks are rejected at Required().
func NewConfirmationPolicy(thresholds map[string]int) ConfirmationPolicy {
	cp := ConfirmationPolicy{required: make(map[string]int, len(thresholds))}
	for network, n := range thresholds {
		network = strings.TrimSpace(network)
		if network == "" || n < 1 {
			continue
		}
		cp.required[network] = n
	}
	return cp
}

// Configured reports whether any thresholds were loaded.
func (p ConfirmationPolicy) Configured() bool {
	return len(p.required) > 0
}

// Required returns the minimum confirmations for network.
func (p ConfirmationPolicy) Required(network string) (int, error) {
	n, ok := p.required[strings.TrimSpace(network)]
	if !ok {
		return 0, fmt.Errorf("%w: no confirmation threshold for %q", ErrInvalid, network)
	}
	return n, nil
}
