package adapter

import (
	"log/slog"

	"github.com/payment-system/dq-vault/lib/adapter/evm"
)

var i *AdapterInventory

func GetInventory(logger *slog.Logger) *AdapterInventory {
	if i == nil {
		i = NewAdapterInventory(
			logger,
			evm.NewEthereumAdapter(logger),
		)
	}
	return i
}