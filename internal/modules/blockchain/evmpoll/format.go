package evmpoll

import (
	"fmt"
	"math/big"
	"strings"
)

var weiPerEth = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

// FormatWeiAsETH converts wei to a decimal ETH string without scientific notation.
// Trailing zeros after the decimal point are trimmed ("1.5000" → "1.5", "1.0" → "1").
func FormatWeiAsETH(wei *big.Int) (string, error) {
	if wei == nil {
		return "", fmt.Errorf("wei is nil")
	}
	if wei.Sign() < 0 {
		return "", fmt.Errorf("wei must be non-negative")
	}
	rat := new(big.Rat).SetFrac(new(big.Int).Set(wei), weiPerEth)
	s := rat.FloatString(18)
	if !strings.Contains(s, ".") {
		return s, nil
	}
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s, nil
}

// Confirmations returns tip-distance confirmations (inclusive of the inclusion block).
func Confirmations(tip, inclusion uint64) int {
	if tip < inclusion {
		return 0
	}
	return int(tip - inclusion + 1)
}

// NextCursorRange returns the inclusive [from, to] block range to scan.
// from = cursor+1; to = min(cursor+batch, tip). ok=false when caught up.
func NextCursorRange(cursor, tip uint64, batch int) (from, to uint64, ok bool) {
	if batch < 1 {
		batch = 1
	}
	if cursor >= tip {
		return 0, 0, false
	}
	from = cursor + 1
	to = cursor + uint64(batch)
	if to > tip {
		to = tip
	}
	return from, to, true
}

// InitCursor picks the starting cursor when none is stored.
// startBlock>0 → startBlock-1 (so first poll begins at startBlock);
// otherwise tip-1 (or 0 if tip is 0) to avoid full history replay on first boot.
func InitCursor(startBlock, tip uint64) uint64 {
	if startBlock > 0 {
		return startBlock - 1
	}
	if tip == 0 {
		return 0
	}
	return tip - 1
}
