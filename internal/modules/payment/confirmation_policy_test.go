package payment_test

import (
	"errors"
	"testing"

	"github.com/DryundeL/crypto-pay/internal/modules/payment"
)

func TestConfirmationPolicyRequired(t *testing.T) {
	p := payment.NewConfirmationPolicy(map[string]int{
		"evm:sepolia": 1,
		"btc:regtest": 1,
	})
	if !p.Configured() {
		t.Fatal("expected configured policy")
	}
	n, err := p.Required("evm:sepolia")
	if err != nil || n != 1 {
		t.Fatalf("Required = %d, %v", n, err)
	}
	_, err = p.Required("evm:mainnet")
	if !errors.Is(err, payment.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}
