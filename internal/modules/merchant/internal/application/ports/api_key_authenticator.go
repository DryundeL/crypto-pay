package ports

import "context"

// AuthPrincipal is the result of a successful API key authentication.
type AuthPrincipal struct {
	MerchantID string
	APIKeyID   string
}

// APIKeyAuthenticator resolves an active API key by its hash.
type APIKeyAuthenticator interface {
	AuthenticateByHash(ctx context.Context, keyHash string) (AuthPrincipal, error)
}
