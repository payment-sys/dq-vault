package adapter

import "log/slog"

type adapter interface {
	CanDo(coinType uint16) bool
	DerivePrivateKey(seed []byte, derivationPath string, isDev bool) (string, error)
	DerivePublicKey(seed []byte, derivationPath string, isDev bool) (string, error)
	DeriveAddress(seed []byte, derivationPath string, isDev bool) (string, error)
	CreateSignedTransaction(seed []byte, derivationPath string, payload string) (string, error)
}

type Inventory struct {
	logger   *slog.Logger
	adapters []adapter
}

func NewAdapterInventory(logger *slog.Logger, adapters ...adapter) *Inventory {
	return &Inventory{
		logger:   logger,
		adapters: adapters,
	}
}

func (i *Inventory) getProvider(coinType uint16) adapter {
	for _, adapter := range i.adapters {
		if adapter.CanDo(coinType) {
			return adapter
		}
	}
	return nil
}

func (i *Inventory) DerivePublicKey(seed []byte, coinType uint16,
	derivationPath string, isDev bool) (string, error) {
	logger := i.logger.With(slog.String("op", "derive_public_key"), slog.Uint64("coinType", uint64(coinType)))
	logger.Info("Deriving public key")

	adapter := i.getProvider(coinType)
	if adapter == nil {
		logger.Error("No adapter found for coin type", "coinType", coinType)
		return "", ErrNoAdapterFound
	}

	pubKey, err := adapter.DerivePublicKey(seed, derivationPath, isDev)
	if err != nil {
		logger.Error("Failed to derive public key", "error", err)
		return "", err
	}

	logger.Info("Public key derived successfully", "pubKey", pubKey)

	return pubKey, nil
}

func (i *Inventory) DeriveAddress(seed []byte, coinType uint16,
	derivationPath string, isDev bool) (string, error) {
	logger := i.logger.With(slog.String("op", "derive_address"), slog.Uint64("coinType", uint64(coinType)))
	logger.Info("Deriving address")

	adapter := i.getProvider(coinType)
	if adapter == nil {
		logger.Error("No adapter found for coin type", "coinType", coinType)
		return "", ErrNoAdapterFound
	}

	address, err := adapter.DeriveAddress(seed, derivationPath, isDev)
	if err != nil {
		logger.Error("Failed to derive address", "error", err)
		return "", err
	}

	logger.Info("Address derived successfully", "address", address)

	return address, nil
}

func (i *Inventory) CreateSignedTransaction(seed []byte, coinType uint16,
	derivationPath string, payload string, _ bool) (string, error) {
	logger := i.logger.With(slog.String("op", "create_signed_transaction"), slog.Uint64("coinType", uint64(coinType)))
	logger.Info("Creating signed transaction")

	adapter := i.getProvider(coinType)
	if adapter == nil {
		logger.Error("No adapter found for coin type", "coinType", coinType)
		return "", ErrNoAdapterFound
	}

	tx, err := adapter.CreateSignedTransaction(seed, derivationPath, payload)
	if err != nil {
		logger.Error("Failed to create signed transaction", "error", err)
		return "", err
	}

	logger.Info("Signed transaction created successfully", "tx", tx)

	return tx, nil
}
