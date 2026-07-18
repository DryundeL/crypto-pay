package authctx

import "context"

type contextKey string

const (
	ctxMerchantID contextKey = "merchant.merchant_id"
	ctxAPIKeyID   contextKey = "merchant.api_key_id"
)

func WithAuth(ctx context.Context, merchantID, apiKeyID string) context.Context {
	ctx = context.WithValue(ctx, ctxMerchantID, merchantID)
	ctx = context.WithValue(ctx, ctxAPIKeyID, apiKeyID)
	return ctx
}

func MerchantID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxMerchantID).(string)
	return v, ok && v != ""
}

func APIKeyID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxAPIKeyID).(string)
	return v, ok && v != ""
}
