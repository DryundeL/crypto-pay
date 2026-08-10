package evmpoll_test

import (
	"math/big"
	"testing"

	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/evmpoll"
)

func TestFormatWeiAsETH(t *testing.T) {
	cases := []struct {
		wei  string
		want string
	}{
		{"0", "0"},
		{"1000000000000000000", "1"},
		{"1500000000000000000", "1.5"},
		{"10000000000000000", "0.01"},
		{"1", "0.000000000000000001"},
	}
	for _, tc := range cases {
		wei, ok := new(big.Int).SetString(tc.wei, 10)
		if !ok {
			t.Fatalf("bad wei %s", tc.wei)
		}
		got, err := evmpoll.FormatWeiAsETH(wei)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("wei %s: got %s want %s", tc.wei, got, tc.want)
		}
	}
}

func TestConfirmations(t *testing.T) {
	if got := evmpoll.Confirmations(100, 100); got != 1 {
		t.Fatalf("got %d", got)
	}
	if got := evmpoll.Confirmations(101, 100); got != 2 {
		t.Fatalf("got %d", got)
	}
	if got := evmpoll.Confirmations(99, 100); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestNextCursorRange(t *testing.T) {
	from, to, ok := evmpoll.NextCursorRange(10, 100, 20)
	if !ok || from != 11 || to != 30 {
		t.Fatalf("got %d-%d ok=%v", from, to, ok)
	}
	from, to, ok = evmpoll.NextCursorRange(95, 100, 20)
	if !ok || from != 96 || to != 100 {
		t.Fatalf("got %d-%d ok=%v", from, to, ok)
	}
	_, _, ok = evmpoll.NextCursorRange(100, 100, 20)
	if ok {
		t.Fatal("expected caught up")
	}
}

func TestInitCursor(t *testing.T) {
	if got := evmpoll.InitCursor(0, 50); got != 49 {
		t.Fatalf("got %d", got)
	}
	if got := evmpoll.InitCursor(40, 50); got != 39 {
		t.Fatalf("got %d", got)
	}
	if got := evmpoll.InitCursor(0, 0); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestMatchNativeTransfers(t *testing.T) {
	watched := evmpoll.WatchSet([]string{"0xAbc", "0xDEF"})
	oneEth := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	txs := []evmpoll.TxView{
		{Hash: "0x1", To: "0xabc", Value: oneEth},
		{Hash: "0x2", To: "0x999", Value: oneEth},
		{Hash: "0x3", To: "0xdef", Value: big.NewInt(0)},
		{Hash: "0x4", To: "", Value: oneEth},
	}
	got, err := evmpoll.MatchNativeTransfers(watched, 42, txs)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("len=%d want 1", len(got))
	}
	if got[0].ToAddress != "0xabc" || got[0].AmountETH != "1" || got[0].BlockNum != 42 {
		t.Fatalf("%+v", got[0])
	}
}
