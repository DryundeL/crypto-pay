package read

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant/internal/domain"
)

type APIKeyAuthenticator struct {
	db *gorm.DB
}

func NewAPIKeyAuthenticator(db *gorm.DB) *APIKeyAuthenticator {
	return &APIKeyAuthenticator{db: db}
}

type authRowRow struct {
	APIKeyID     string `gorm:"column:api_key_id"`
	MerchantID   string `gorm:"column:merchant_id"`
	KeyStatus    string `gorm:"column:key_status"`
	MerchantStat string `gorm:"column:merchant_status"`
}

func (a *APIKeyAuthenticator) AuthenticateByHash(ctx context.Context, keyHash string) (ports.AuthPrincipal, error) {
	if keyHash == "" {
		return ports.AuthPrincipal{}, domain.ErrInvalidAPIKey
	}

	var row authRowRow
	err := a.db.WithContext(ctx).
		Table("merchant_api_keys AS k").
		Select("k.id AS api_key_id, k.merchant_id, k.status AS key_status, m.status AS merchant_status").
		Joins("INNER JOIN merchants AS m ON m.id = k.merchant_id").
		Where("k.key_hash = ?", keyHash).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ports.AuthPrincipal{}, domain.ErrAPIKeyNotFound
		}
		return ports.AuthPrincipal{}, fmt.Errorf("authenticate api key: %w", err)
	}

	if row.KeyStatus != string(domain.APIKeyStatusActive) {
		return ports.AuthPrincipal{}, domain.ErrAPIKeyRevoked
	}
	if row.MerchantStat != string(domain.StatusActive) {
		return ports.AuthPrincipal{}, domain.ErrMerchantSuspended
	}

	return ports.AuthPrincipal{
		MerchantID: row.MerchantID,
		APIKeyID:   row.APIKeyID,
	}, nil
}
