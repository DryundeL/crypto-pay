package domain

import (
	"fmt"
	"time"
)

type APIKeyID string

func (id APIKeyID) String() string { return string(id) }

// APIKey is an entity owned by Merchant aggregate.
// Plaintext secret is never stored on the entity — only hash + prefix.
type APIKey struct {
	id        APIKeyID
	name      string
	keyPrefix string
	keyHash   string
	status    APIKeyStatus
	createdAt time.Time
	revokedAt *time.Time
}

func restoreAPIKey(
	id APIKeyID,
	name, keyPrefix, keyHash string,
	status APIKeyStatus,
	createdAt time.Time,
	revokedAt *time.Time,
) *APIKey {
	return &APIKey{
		id:        id,
		name:      name,
		keyPrefix: keyPrefix,
		keyHash:   keyHash,
		status:    status,
		createdAt: createdAt,
		revokedAt: revokedAt,
	}
}

func newAPIKey(id APIKeyID, name, keyPrefix, keyHash string, now time.Time) (*APIKey, error) {
	if id.String() == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidAPIKey)
	}
	if name == "" || len(name) > 100 {
		return nil, fmt.Errorf("%w: name must be 1..100 chars", ErrInvalidAPIKey)
	}
	if keyPrefix == "" || keyHash == "" {
		return nil, fmt.Errorf("%w: key material is required", ErrInvalidAPIKey)
	}
	return &APIKey{
		id:        id,
		name:      name,
		keyPrefix: keyPrefix,
		keyHash:   keyHash,
		status:    APIKeyStatusActive,
		createdAt: now.UTC(),
	}, nil
}

func (k *APIKey) ID() APIKeyID          { return k.id }
func (k *APIKey) Name() string          { return k.name }
func (k *APIKey) KeyPrefix() string     { return k.keyPrefix }
func (k *APIKey) KeyHash() string       { return k.keyHash }
func (k *APIKey) Status() APIKeyStatus  { return k.status }
func (k *APIKey) CreatedAt() time.Time  { return k.createdAt }
func (k *APIKey) RevokedAt() *time.Time { return k.revokedAt }
func (k *APIKey) IsActive() bool        { return k.status == APIKeyStatusActive }

func (k *APIKey) revoke(now time.Time) error {
	if k.status == APIKeyStatusRevoked {
		return ErrAPIKeyRevoked
	}
	ts := now.UTC()
	k.status = APIKeyStatusRevoked
	k.revokedAt = &ts
	return nil
}
