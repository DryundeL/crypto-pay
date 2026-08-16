package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

func TestGenerateWebhookSecretIsUnique(t *testing.T) {
	a, err := domain.GenerateWebhookSecret()
	if err != nil {
		t.Fatal(err)
	}
	b, err := domain.GenerateWebhookSecret()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("two merchants must not share a webhook secret")
	}
	if len(a) < 32 || len(b) < 32 {
		t.Fatalf("secrets too short: %d %d", len(a), len(b))
	}
}

func TestNewMerchantRequiresWebhookSecret(t *testing.T) {
	_, err := domain.NewMerchant(domain.NewMerchantParams{
		ID:   domain.MerchantID("m-1"),
		Name: "Acme",
		Now:  time.Unix(1, 0).UTC(),
	})
	if !errors.Is(err, domain.ErrInvalidMerchant) {
		t.Fatalf("err = %v", err)
	}
}

func TestRotateWebhookSecret(t *testing.T) {
	first, err := domain.GenerateWebhookSecret()
	if err != nil {
		t.Fatal(err)
	}
	m, err := domain.NewMerchant(domain.NewMerchantParams{
		ID:            domain.MerchantID("m-1"),
		Name:          "Acme",
		WebhookSecret: first,
		Now:           time.Unix(1, 0).UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if m.WebhookSecret() != first {
		t.Fatal("create must persist the generated secret")
	}

	next, err := domain.GenerateWebhookSecret()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.RotateWebhookSecret(next, time.Unix(2, 0).UTC()); err != nil {
		t.Fatal(err)
	}
	if m.WebhookSecret() != next {
		t.Fatalf("secret = %s, want rotated value", m.WebhookSecret())
	}
}
