package merchant

import (
	"context"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/authctx"
)

// WithAuth stores authenticated merchant principal in context.
func WithAuth(ctx context.Context, merchantID, apiKeyID string) context.Context {
	return authctx.WithAuth(ctx, merchantID, apiKeyID)
}

func MerchantIDFromContext(ctx context.Context) (string, bool) {
	return authctx.MerchantID(ctx)
}

func APIKeyIDFromContext(ctx context.Context) (string, bool) {
	return authctx.APIKeyID(ctx)
}
