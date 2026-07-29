package domain_test

import (
	"strings"
	"testing"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
)

func TestDeriveAddress(t *testing.T) {
	a1, path1, err := domain.DeriveAddress(domain.NetworkEVMSepolia, 0, "inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if path1 != "m/0/0" {
		t.Fatalf("path = %s", path1)
	}
	if !strings.HasPrefix(a1, "0x") || len(a1) != 42 {
		t.Fatalf("evm address = %s", a1)
	}

	a2, _, err := domain.DeriveAddress(domain.NetworkEVMSepolia, 0, "inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if a1 != a2 {
		t.Fatal("address must be stable for same inputs")
	}

	a3, _, err := domain.DeriveAddress(domain.NetworkEVMSepolia, 1, "inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if a1 == a3 {
		t.Fatal("different index must change address")
	}

	btc, path, err := domain.DeriveAddress(domain.NetworkBitcoinRegtest, 0, "inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if path != "m/0/0" {
		t.Fatalf("path = %s", path)
	}
	if !strings.HasPrefix(btc, "bcrt1q") {
		t.Fatalf("btc address = %s", btc)
	}

	tb, _, err := domain.DeriveAddress(domain.NetworkBitcoinTestnet, 0, "inv-1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tb, "tb1q") {
		t.Fatalf("testnet address = %s", tb)
	}
}
