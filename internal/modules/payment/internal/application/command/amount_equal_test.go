package command

import "testing"

func TestAmountsEqual(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1", "1.0", true},
		{"1.00", "1", true},
		{"0.01", "0.010", true},
		{"1.5", "1.50", true},
		{"1", "2", false},
		{"", "", true},
		{"abc", "abc", true},
		{"abc", "abd", false},
	}
	for _, tc := range cases {
		if got := amountsEqual(tc.a, tc.b); got != tc.want {
			t.Fatalf("amountsEqual(%q,%q)=%v want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCurrenciesEqual(t *testing.T) {
	if !currenciesEqual("ETH", "eth") {
		t.Fatal("expected case-insensitive match")
	}
	if currenciesEqual("ETH", "BTC") {
		t.Fatal("expected mismatch")
	}
}
