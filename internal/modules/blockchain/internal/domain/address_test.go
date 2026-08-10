package domain_test

import (
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/internal/domain"
)

func testAccountXPub(t *testing.T) string {
	t.Helper()
	seed := make([]byte, 64)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	purpose, err := master.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		t.Fatal(err)
	}
	coin, err := purpose.Derive(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		t.Fatal(err)
	}
	account, err := coin.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		t.Fatal(err)
	}
	neutered, err := account.Neuter()
	if err != nil {
		t.Fatal(err)
	}
	return neutered.String()
}

func TestHDAddressDeriver_EVM(t *testing.T) {
	xpub := testAccountXPub(t)
	d, err := domain.NewHDAddressDeriver(xpub)
	if err != nil {
		t.Fatal(err)
	}

	a1, path1, err := d.Derive(domain.NetworkEVMSepolia, 0)
	if err != nil {
		t.Fatal(err)
	}
	if path1 != "m/44'/60'/0'/0/0" {
		t.Fatalf("path = %s", path1)
	}
	if !strings.HasPrefix(a1, "0x") || len(a1) != 42 {
		t.Fatalf("evm address = %s", a1)
	}
	if a1 != strings.ToLower(a1) {
		t.Fatalf("address must be lowercase: %s", a1)
	}

	a2, _, err := d.Derive(domain.NetworkEVMSepolia, 0)
	if err != nil {
		t.Fatal(err)
	}
	if a1 != a2 {
		t.Fatal("address must be stable for same index")
	}

	a3, path3, err := d.Derive(domain.NetworkEVMSepolia, 1)
	if err != nil {
		t.Fatal(err)
	}
	if a1 == a3 {
		t.Fatal("different index must change address")
	}
	if path3 != "m/44'/60'/0'/0/1" {
		t.Fatalf("path = %s", path3)
	}
}

func TestHDAddressDeriver_EVMRequiresXPub(t *testing.T) {
	d, err := domain.NewHDAddressDeriver("")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = d.Derive(domain.NetworkEVMSepolia, 0)
	if err == nil {
		t.Fatal("expected error without xpub")
	}
}

func TestHDAddressDeriver_BTCPlaceholder(t *testing.T) {
	d, err := domain.NewHDAddressDeriver("")
	if err != nil {
		t.Fatal(err)
	}
	btc, path, err := d.Derive(domain.NetworkBitcoinRegtest, 0)
	if err != nil {
		t.Fatal(err)
	}
	if path != "m/0/0" {
		t.Fatalf("path = %s", path)
	}
	if !strings.HasPrefix(btc, "bcrt1q") {
		t.Fatalf("btc address = %s", btc)
	}

	tb, _, err := d.Derive(domain.NetworkBitcoinTestnet, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tb, "tb1q") {
		t.Fatalf("testnet address = %s", tb)
	}
}

func TestNewHDAddressDeriver_RejectsXPrv(t *testing.T) {
	seed := make([]byte, 64)
	for i := range seed {
		seed[i] = byte(i + 1)
	}
	master, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		t.Fatal(err)
	}
	_, err = domain.NewHDAddressDeriver(master.String())
	if err == nil {
		t.Fatal("expected error for private key")
	}
}
