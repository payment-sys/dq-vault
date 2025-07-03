package adapter

import (
	"log/slog"
	"sync"

	"github.com/payment-system/dq-vault/lib/adapter/evm"
)

// Package-level variables for singleton pattern
var (
	inventory *Inventory //nolint:gochecknoglobals // singleton pattern requires global state
	once      sync.Once  //nolint:gochecknoglobals // singleton pattern requires global state
)

// GetInventory returns the singleton adapter inventory instance
func GetInventory(logger *slog.Logger) *Inventory {
	once.Do(func() {
		inventory = NewAdapterInventory(
			logger,
			evm.NewEthereumAdapter(logger),
		)
	})
	return inventory
}
