package domain

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	webhookSecretPrefix = "whsec_"
	webhookSecretBytes  = 32
	minWebhookSecretLen = 32
)

// Merchant is the aggregate root of the Merchant bounded context.
type Merchant struct {
	id            MerchantID
	name          string
	status        Status
	webhookURL    string
	webhookSecret string
	apiKeys       []*APIKey
	createdAt     time.Time
	updatedAt     time.Time
}

type NewMerchantParams struct {
	ID            MerchantID
	Name          string
	WebhookURL    string
	WebhookSecret string
	Now           time.Time
}

func NewMerchant(p NewMerchantParams) (*Merchant, error) {
	name := strings.TrimSpace(p.Name)
	if p.ID.IsZero() {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidMerchant)
	}
	if name == "" || len(name) > 200 {
		return nil, fmt.Errorf("%w: name must be 1..200 chars", ErrInvalidMerchant)
	}
	webhookURL, err := normalizeWebhookURL(p.WebhookURL)
	if err != nil {
		return nil, err
	}
	webhookSecret, err := normalizeWebhookSecret(p.WebhookSecret)
	if err != nil {
		return nil, err
	}
	now := p.Now.UTC()
	return &Merchant{
		id:            p.ID,
		name:          name,
		status:        StatusActive,
		webhookURL:    webhookURL,
		webhookSecret: webhookSecret,
		apiKeys:       make([]*APIKey, 0),
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

func RestoreMerchant(
	id MerchantID,
	name string,
	status Status,
	webhookURL, webhookSecret string,
	apiKeys []*APIKey,
	createdAt, updatedAt time.Time,
) *Merchant {
	if apiKeys == nil {
		apiKeys = make([]*APIKey, 0)
	}
	return &Merchant{
		id:            id,
		name:          name,
		status:        status,
		webhookURL:    webhookURL,
		webhookSecret: webhookSecret,
		apiKeys:       apiKeys,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (m *Merchant) ID() MerchantID        { return m.id }
func (m *Merchant) Name() string          { return m.name }
func (m *Merchant) Status() Status        { return m.status }
func (m *Merchant) WebhookURL() string    { return m.webhookURL }
func (m *Merchant) WebhookSecret() string { return m.webhookSecret }
func (m *Merchant) APIKeys() []*APIKey    { return append([]*APIKey(nil), m.apiKeys...) }
func (m *Merchant) CreatedAt() time.Time  { return m.createdAt }
func (m *Merchant) UpdatedAt() time.Time  { return m.updatedAt }

func (m *Merchant) IsActive() bool { return m.status == StatusActive }

type IssueAPIKeyParams struct {
	ID        APIKeyID
	Name      string
	KeyPrefix string
	KeyHash   string
	Now       time.Time
}

// IssueAPIKey adds a new active API key to the aggregate.
func (m *Merchant) IssueAPIKey(p IssueAPIKeyParams) (*APIKey, error) {
	if !m.IsActive() {
		return nil, ErrMerchantSuspended
	}
	key, err := newAPIKey(p.ID, strings.TrimSpace(p.Name), p.KeyPrefix, p.KeyHash, p.Now)
	if err != nil {
		return nil, err
	}
	m.apiKeys = append(m.apiKeys, key)
	m.updatedAt = p.Now.UTC()
	return key, nil
}

// RevokeAPIKey revokes an owned API key.
func (m *Merchant) RevokeAPIKey(keyID APIKeyID, now time.Time) (*APIKey, error) {
	for _, key := range m.apiKeys {
		if key.ID() == keyID {
			if err := key.revoke(now); err != nil {
				return nil, err
			}
			m.updatedAt = now.UTC()
			return key, nil
		}
	}
	return nil, ErrAPIKeyNotFound
}

// RotateWebhookSecret replaces the HMAC signing secret. Plaintext is returned only by the command.
func (m *Merchant) RotateWebhookSecret(secret string, now time.Time) error {
	if !m.IsActive() {
		return ErrMerchantSuspended
	}
	normalized, err := normalizeWebhookSecret(secret)
	if err != nil {
		return err
	}
	m.webhookSecret = normalized
	m.updatedAt = now.UTC()
	return nil
}

// GenerateWebhookSecret returns a unique HMAC key (whsec_ + 32 random bytes, base64url).
func GenerateWebhookSecret() (string, error) {
	buf := make([]byte, webhookSecretBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate webhook secret: %w", err)
	}
	return webhookSecretPrefix + base64.RawURLEncoding.EncodeToString(buf), nil
}

func normalizeWebhookSecret(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if len(raw) < minWebhookSecretLen {
		return "", fmt.Errorf("%w: webhook_secret must be at least %d characters", ErrInvalidMerchant, minWebhookSecretLen)
	}
	return raw, nil
}

func normalizeWebhookURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("%w: webhook_url must be absolute http(s) URL", ErrInvalidMerchant)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("%w: webhook_url scheme must be http or https", ErrInvalidMerchant)
	}
	return u.String(), nil
}

// RestoreAPIKey rebuilds an API key entity from persistence (package-local factory for write repo).
func RestoreAPIKey(
	id APIKeyID,
	name, keyPrefix, keyHash string,
	status APIKeyStatus,
	createdAt time.Time,
	revokedAt *time.Time,
) *APIKey {
	return restoreAPIKey(id, name, keyPrefix, keyHash, status, createdAt, revokedAt)
}
