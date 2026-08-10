package app

import (
	"fmt"
	"strconv"
	"strings"
)

// DefaultConfirmationThresholds returns per-network minimum confirmations.
// Development/staging stay at 1 for local chains; production uses safer floors.
func DefaultConfirmationThresholds(env string) map[string]int {
	switch env {
	case "production":
		return map[string]int{
			"evm:sepolia": 12,
			"btc:regtest": 6,
			"btc:testnet": 3,
		}
	default: // development, staging
		return map[string]int{
			"evm:sepolia": 1,
			"btc:regtest": 1,
			"btc:testnet": 1,
		}
	}
}

// ParseConfirmationThresholds parses "network=n,network=n" overrides.
func ParseConfirmationThresholds(raw string) (map[string]int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	out := make(map[string]int)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		network, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid confirmation threshold %q (want network=N)", part)
		}
		network = strings.TrimSpace(network)
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || n < 1 {
			return nil, fmt.Errorf("invalid confirmation threshold %q: N must be >= 1", part)
		}
		if network == "" {
			return nil, fmt.Errorf("invalid confirmation threshold %q: empty network", part)
		}
		out[network] = n
	}
	return out, nil
}

func mergeConfirmationThresholds(base, override map[string]int) map[string]int {
	out := make(map[string]int, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}
