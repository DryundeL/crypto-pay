package app

import (
	"testing"
)

func TestDefaultConfirmationThresholds(t *testing.T) {
	dev := DefaultConfirmationThresholds("development")
	if dev["evm:sepolia"] != 1 || dev["btc:regtest"] != 1 {
		t.Fatalf("dev thresholds = %#v", dev)
	}
	prod := DefaultConfirmationThresholds("production")
	if prod["evm:sepolia"] < 2 {
		t.Fatalf("prod sepolia must be stricter than 1, got %d", prod["evm:sepolia"])
	}
}

func TestParseConfirmationThresholds(t *testing.T) {
	got, err := ParseConfirmationThresholds("evm:sepolia=2, btc:regtest=3")
	if err != nil {
		t.Fatal(err)
	}
	if got["evm:sepolia"] != 2 || got["btc:regtest"] != 3 {
		t.Fatalf("got %#v", got)
	}

	merged := mergeConfirmationThresholds(DefaultConfirmationThresholds("development"), got)
	if merged["evm:sepolia"] != 2 || merged["btc:testnet"] != 1 {
		t.Fatalf("merged %#v", merged)
	}

	if _, err := ParseConfirmationThresholds("evm:sepolia=0"); err == nil {
		t.Fatal("expected error for N < 1")
	}
}
