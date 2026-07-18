package write

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
)

const (
	apiKeyPrefixTag = "cpay_"
	apiKeyBytes     = 32
	displayPrefixN  = 12
)

// APIKeyGenerator creates cryptographically random API keys.
// Hash = SHA-256(pepper || ":" || plaintext). Plaintext is never persisted.
type APIKeyGenerator struct {
	pepper string
}

func NewAPIKeyGenerator(pepper string) *APIKeyGenerator {
	return &APIKeyGenerator{pepper: pepper}
}

func (g *APIKeyGenerator) Generate() (ports.APIKeyMaterial, error) {
	if g.pepper == "" {
		return ports.APIKeyMaterial{}, fmt.Errorf("api key pepper is required")
	}

	buf := make([]byte, apiKeyBytes)
	if _, err := rand.Read(buf); err != nil {
		return ports.APIKeyMaterial{}, fmt.Errorf("read random: %w", err)
	}

	secret := base64.RawURLEncoding.EncodeToString(buf)
	plaintext := apiKeyPrefixTag + secret
	prefix := plaintext
	if len(prefix) > displayPrefixN {
		prefix = prefix[:displayPrefixN]
	}

	sum := sha256.Sum256([]byte(g.pepper + ":" + plaintext))
	hash := hex.EncodeToString(sum[:])

	return ports.APIKeyMaterial{
		Plaintext: plaintext,
		Prefix:    prefix,
		Hash:      hash,
	}, nil
}

// Hash hashes an inbound API key for lookup/auth against merchant_api_keys.key_hash.
func (g *APIKeyGenerator) Hash(plaintext string) string {
	plaintext = strings.TrimSpace(plaintext)
	sum := sha256.Sum256([]byte(g.pepper + ":" + plaintext))
	return hex.EncodeToString(sum[:])
}
