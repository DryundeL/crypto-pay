package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/application/ports"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger/internal/domain"
)

type PostDebitJournalHandler struct {
	accounts domain.AccountRepository
	journals domain.JournalRepository
	tx       ports.TransactionManager
	pub      ports.EventPublisher
	now      func() time.Time
	newID    func() string
}

func NewPostDebitJournalHandler(
	accounts domain.AccountRepository,
	journals domain.JournalRepository,
	tx ports.TransactionManager,
	pub ports.EventPublisher,
) *PostDebitJournalHandler {
	return &PostDebitJournalHandler{
		accounts: accounts,
		journals: journals,
		tx:       tx,
		pub:      pub,
		now:      time.Now,
		newID:    func() string { return uuid.NewString() },
	}
}

func (h *PostDebitJournalHandler) Handle(ctx context.Context, cmd PostDebitJournal) (PostDebitJournalResult, error) {
	money, err := domain.NewMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return PostDebitJournalResult{}, err
	}

	var result PostDebitJournalResult
	err = h.tx.WithinTransaction(ctx, func(ctx context.Context) error {
		existing, err := h.journals.FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
		if err == nil {
			result = PostDebitJournalResult{JournalID: existing.ID().String(), Created: false}
			return nil
		}
		if !errors.Is(err, domain.ErrJournalNotFound) {
			return err
		}

		now := h.now().UTC()
		clearing, err := h.ensureAccount(ctx, domain.OwnerSystem, domain.SystemOwnerID, domain.KindClearing, money.Currency(), now)
		if err != nil {
			return err
		}
		merchantAcc, err := h.ensureAccount(ctx, domain.OwnerMerchant, cmd.MerchantID, domain.KindAvailable, money.Currency(), now)
		if err != nil {
			return err
		}

		journal, err := domain.PostDebit(domain.PostParams{
			ID:                domain.JournalID(h.newID()),
			IdempotencyKey:    cmd.IdempotencyKey,
			MerchantID:        cmd.MerchantID,
			Amount:            money,
			ReferenceType:     cmd.ReferenceType,
			ReferenceID:       cmd.ReferenceID,
			ClearingAccountID: clearing.ID(),
			MerchantAccountID: merchantAcc.ID(),
			ClearingEntryID:   domain.EntryID(h.newID()),
			MerchantEntryID:   domain.EntryID(h.newID()),
			Now:               now,
		})
		if err != nil {
			return err
		}

		if err := merchantAcc.ApplyEntry(domain.SideDebit, money, now); err != nil {
			return err
		}
		if err := clearing.ApplyEntry(domain.SideCredit, money, now); err != nil {
			return err
		}
		if err := h.accounts.Save(ctx, clearing); err != nil {
			return fmt.Errorf("save clearing account: %w", err)
		}
		if err := h.accounts.Save(ctx, merchantAcc); err != nil {
			return fmt.Errorf("save merchant account: %w", err)
		}
		if err := h.journals.Save(ctx, journal); err != nil {
			return fmt.Errorf("save journal: %w", err)
		}

		if err := h.pub.Publish(ctx, "ledger", journal.ID().String(), entryPostedEvent{
			JournalID:     journal.ID().String(),
			MerchantID:    journal.MerchantID(),
			Amount:        money.Amount(),
			Currency:      money.Currency(),
			ReferenceType: journal.ReferenceType(),
			ReferenceID:   journal.ReferenceID(),
			OccurredAt:    now,
		}); err != nil {
			return err
		}

		result = PostDebitJournalResult{JournalID: journal.ID().String(), Created: true}
		return nil
	})
	return result, err
}

func (h *PostDebitJournalHandler) ensureAccount(
	ctx context.Context,
	ownerType domain.OwnerType,
	ownerID string,
	kind domain.AccountKind,
	currency string,
	now time.Time,
) (*domain.Account, error) {
	acc, err := h.accounts.FindByKey(ctx, ownerType, ownerID, kind, currency)
	if err == nil {
		return acc, nil
	}
	if !errors.Is(err, domain.ErrAccountNotFound) {
		return nil, err
	}

	acc, err = domain.CreateAccount(domain.CreateAccountParams{
		ID:        domain.AccountID(h.newID()),
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Kind:      kind,
		Currency:  currency,
		Now:       now,
	})
	if err != nil {
		return nil, err
	}
	if err := h.accounts.Save(ctx, acc); err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return acc, nil
}
