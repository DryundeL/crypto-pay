package evmpoll

import (
	"math/big"
	"strings"
)

// NativeTransfer is a matched native ETH payment candidate.
type NativeTransfer struct {
	TxHash    string
	ToAddress string
	AmountETH string
	Currency  string
	BlockNum  uint64
}

// MatchNativeTransfers filters txs whose To is in watched (lowercase) and value > 0.
func MatchNativeTransfers(
	watched map[string]struct{},
	blockNum uint64,
	txs []TxView,
) ([]NativeTransfer, error) {
	if len(watched) == 0 || len(txs) == 0 {
		return nil, nil
	}
	out := make([]NativeTransfer, 0)
	for _, tx := range txs {
		if tx.To == "" || tx.Value == nil || tx.Value.Sign() <= 0 {
			continue
		}
		to := strings.ToLower(tx.To)
		if _, ok := watched[to]; !ok {
			continue
		}
		amount, err := FormatWeiAsETH(tx.Value)
		if err != nil {
			return nil, err
		}
		out = append(out, NativeTransfer{
			TxHash:    strings.ToLower(tx.Hash),
			ToAddress: to,
			AmountETH: amount,
			Currency:  "ETH",
			BlockNum:  blockNum,
		})
	}
	return out, nil
}

// TxView is a minimal transaction projection for matching (keeps tests RPC-free).
type TxView struct {
	Hash  string
	To    string
	Value *big.Int
}

// WatchSet builds a lowercase address set from a list.
func WatchSet(addresses []string) map[string]struct{} {
	out := make(map[string]struct{}, len(addresses))
	for _, a := range addresses {
		a = strings.ToLower(strings.TrimSpace(a))
		if a == "" {
			continue
		}
		out[a] = struct{}{}
	}
	return out
}
