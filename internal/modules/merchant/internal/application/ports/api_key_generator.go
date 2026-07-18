package ports

// APIKeyMaterial is generated secret material for a new API key.
type APIKeyMaterial struct {
	Plaintext string
	Prefix    string
	Hash      string
}

// APIKeyGenerator creates secure API key material (plaintext shown once to the client).
type APIKeyGenerator interface {
	Generate() (APIKeyMaterial, error)
	Hash(plaintext string) string
}
