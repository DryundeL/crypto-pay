package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/DryundeL/crypto-pay/internal/app"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain"
	"github.com/DryundeL/crypto-pay/internal/modules/blockchain/evmpoll"
	"github.com/DryundeL/crypto-pay/internal/modules/invoice"
	"github.com/DryundeL/crypto-pay/internal/modules/ledger"
	"github.com/DryundeL/crypto-pay/internal/modules/merchant"
	"github.com/DryundeL/crypto-pay/internal/modules/payment"
	"github.com/DryundeL/crypto-pay/internal/modules/webhook"
	"github.com/DryundeL/crypto-pay/internal/platform/database"
	"github.com/DryundeL/crypto-pay/internal/platform/observability"
)

const network = string(blockchain.NetworkEVMSepolia)

func main() {
	log := observability.Logger()

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Error("load config failed", "error", err)
		os.Exit(1)
	}
	if strings.TrimSpace(cfg.Scanner.SepoliaRPCURL) == "" {
		log.Error("EVM_SEPOLIA_RPC_URL is required for scanner")
		os.Exit(1)
	}
	if strings.TrimSpace(cfg.EVM.SepoliaXPub) == "" {
		log.Error("EVM_SEPOLIA_XPUB is required for scanner (must match API address derivation)")
		os.Exit(1)
	}

	db, err := database.Connect(cfg.Database.DSN())
	if err != nil {
		log.Error("connect database failed", "error", err)
		os.Exit(1)
	}

	if err := database.RunMigrationsUp(cfg.Database.DSN(), database.LogIdentity{
		User:   cfg.Database.User,
		Host:   cfg.Database.Host,
		Port:   cfg.Database.Port,
		DBName: cfg.Database.DBName,
	}, "migrations"); err != nil {
		log.Error("run migrations failed", "error", err)
		os.Exit(1)
	}

	confirmations := payment.NewConfirmationPolicy(cfg.Confirmations)
	blockchainModule := blockchain.NewModule(blockchain.Dependencies{
		DB:             db,
		Confirmations:  confirmations,
		EVMSepoliaXPub: cfg.EVM.SepoliaXPub,
	})
	invoiceModule := invoice.NewModule(invoice.Dependencies{
		DB:              db,
		AddressProvider: blockchainModule.AddressProvider(),
	})
	paymentModule := payment.NewModule(payment.Dependencies{
		DB:               db,
		InvoiceLookup:    invoiceModule,
		InvoiceLifecycle: invoiceModule,
		Confirmations:    confirmations,
	})
	blockchainModule.SetPaymentNotifier(paymentModule)

	ledgerModule := ledger.NewModule(ledger.Dependencies{DB: db})
	merchantModule := merchant.NewModule(merchant.Dependencies{
		DB:           db,
		APIKeyPepper: cfg.App.JWTSecret,
	})
	webhookModule := webhook.NewModule(webhook.Dependencies{
		DB:            db,
		Merchant:      merchantModule,
		SigningSecret: cfg.App.JWTSecret,
	})
	invoiceModule.SetPaidNotifier(app.NewSideEffectNotifier(ledgerModule, webhookModule))

	client, err := ethclient.Dial(cfg.Scanner.SepoliaRPCURL)
	if err != nil {
		log.Error("dial eth rpc failed", "error", err)
		os.Exit(1)
	}
	defer client.Close()

	required, err := confirmations.Required(network)
	if err != nil {
		log.Error("confirmation policy failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("scanner starting",
		"env", cfg.App.Env,
		"network", network,
		"poll_interval", cfg.Scanner.PollInterval.String(),
		"batch", cfg.Scanner.BlockBatch,
		"required_confirmations", required,
	)

	ticker := time.NewTicker(cfg.Scanner.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("scanner shutting down")
			sqlDB, err := db.DB()
			if err == nil {
				_ = sqlDB.Close()
			}
			return
		case <-ticker.C:
			if err := pollOnce(ctx, client, blockchainModule, confirmations, cfg); err != nil {
				log.Error("scanner tick failed", "error", err)
			}
		}
	}
}

func pollOnce(
	ctx context.Context,
	client *ethclient.Client,
	bc *blockchain.Module,
	policy payment.ConfirmationPolicy,
	cfg *app.Config,
) error {
	log := observability.Logger()

	tip, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("block number: %w", err)
	}

	cursor, found, err := bc.GetScanCursor(ctx, network)
	if err != nil {
		return fmt.Errorf("get cursor: %w", err)
	}
	if !found {
		// Init against lagged tip so cursor does not sit ahead of safeTip when required > 1.
		safeTip := tip
		if req, err := policy.Required(network); err == nil && req > 1 {
			lag := uint64(req - 1)
			if tip >= lag {
				safeTip = tip - lag
			}
		}
		cursor = evmpoll.InitCursor(cfg.Scanner.SepoliaStartBlock, safeTip)
		if err := bc.SaveScanCursor(ctx, network, cursor); err != nil {
			return fmt.Errorf("init cursor: %w", err)
		}
		log.Info("initialized scan cursor", "network", network, "block", cursor, "tip", tip)
	}

	required, err := policy.Required(network)
	if err != nil {
		return err
	}

	// Lag the scan tip so a tx is first seen only when confirmations >= required.
	// Avoids needing to revisit past blocks after the cursor advances (deep reorg later).
	safeTip := tip
	if required > 1 {
		lag := uint64(required - 1)
		if tip < lag {
			return nil
		}
		safeTip = tip - lag
	}

	from, to, ok := evmpoll.NextCursorRange(cursor, safeTip, cfg.Scanner.BlockBatch)
	if !ok {
		return nil
	}

	addrs, err := bc.ListWatchedAddresses(ctx, network)
	if err != nil {
		return fmt.Errorf("list watched: %w", err)
	}
	watchList := make([]string, 0, len(addrs))
	for _, a := range addrs {
		watchList = append(watchList, a.Address)
	}
	watched := evmpoll.WatchSet(watchList)

	for n := from; n <= to; n++ {
		if err := processBlock(ctx, client, bc, watched, n, tip, required); err != nil {
			return fmt.Errorf("block %d: %w", n, err)
		}
		if err := bc.SaveScanCursor(ctx, network, n); err != nil {
			return fmt.Errorf("save cursor %d: %w", n, err)
		}
	}

	log.Info("scanned blocks", "from", from, "to", to, "tip", tip, "safe_tip", safeTip, "watched", len(watched))
	return nil
}

func processBlock(
	ctx context.Context,
	client *ethclient.Client,
	bc *blockchain.Module,
	watched map[string]struct{},
	blockNum, tip uint64,
	required int,
) error {
	block, err := client.BlockByNumber(ctx, new(big.Int).SetUint64(blockNum))
	if err != nil {
		return fmt.Errorf("get block: %w", err)
	}

	txs := make([]evmpoll.TxView, 0, len(block.Transactions()))
	for _, tx := range block.Transactions() {
		txs = append(txs, toTxView(tx))
	}

	matches, err := evmpoll.MatchNativeTransfers(watched, blockNum, txs)
	if err != nil {
		return err
	}

	confs := evmpoll.Confirmations(tip, blockNum)
	for _, m := range matches {
		ref, err := bc.LookupDeposit(ctx, network, m.ToAddress)
		if err != nil {
			if errors.Is(err, blockchain.ErrNotFound) {
				continue
			}
			return fmt.Errorf("lookup %s: %w", m.ToAddress, err)
		}
		if _, err := bc.RecordObservation(ctx, blockchain.RecordObservationInput{
			MerchantID:    ref.MerchantID,
			Network:       network,
			TxHash:        m.TxHash,
			ToAddress:     m.ToAddress,
			Amount:        m.AmountETH,
			Currency:      ref.Currency,
			Confirmations: confs,
		}); err != nil {
			return fmt.Errorf("observe %s: %w", m.TxHash, err)
		}
		if confs < required {
			continue
		}
		if _, err := bc.ConfirmTransaction(ctx, blockchain.ConfirmTransactionInput{
			MerchantID: ref.MerchantID,
			Network:    network,
			TxHash:     m.TxHash,
		}); err != nil {
			if errors.Is(err, blockchain.ErrInsufficientConfirmations) {
				continue
			}
			return fmt.Errorf("confirm %s: %w", m.TxHash, err)
		}
	}
	return nil
}

func toTxView(tx *types.Transaction) evmpoll.TxView {
	view := evmpoll.TxView{
		Hash:  strings.ToLower(tx.Hash().Hex()),
		Value: new(big.Int).Set(tx.Value()),
	}
	if to := tx.To(); to != nil {
		view.To = strings.ToLower(to.Hex())
	}
	return view
}
