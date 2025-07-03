package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/payment-system/dq-vault/api/helpers"
	"github.com/payment-system/dq-vault/config"
	"github.com/payment-system/dq-vault/lib"
	"github.com/payment-system/dq-vault/lib/adapter"
	"github.com/payment-system/dq-vault/lib/slip44"
)

func (b *backend) pathAddress(ctx context.Context, req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	backendLogger := b.logger.With(slog.String("op", "path_address"))
	if err := helpers.ValidateFields(req, d); err != nil {
		backendLogger.Error("validate fields", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	// UUID of user required to sign transaction
	uuid := d.Get("uuid").(string)

	// derivation path
	derivationPath := d.Get("path").(string)

	// coin type of transaction
	// see supported coinTypes lib/bipp44coins
	coinType := d.Get("coinType").(int)

	isDev := d.Get("isDev").(bool)

	if uint16(coinType) == slip44.Bitshares {
		derivationPath = config.BitsharesDerivationPath
	}

	backendLogger.Info("request", "path", derivationPath, "cointype", coinType)

	// validate data provided
	if err := helpers.ValidateData(ctx, req, uuid, derivationPath); err != nil {
		backendLogger.Error("validate data", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	// path where user data is stored in vault
	path := config.StorageBasePath + uuid
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		backendLogger.Error("get", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	// obtain mnemonic and passphrase of user
	var userInfo helpers.User
	err = entry.DecodeJSON(&userInfo)
	if err != nil {
		backendLogger.Error("decode json", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	seed, err := lib.SeedFromMnemonic(userInfo.Mnemonic, userInfo.Passphrase)

	backendLogger.Info("dp", "dp", derivationPath)

	// obtains blockchain adapater based on coinType
	adapter := adapter.GetInventory(backendLogger)
	if err != nil {
		backendLogger.Error("get adapter", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	address, err := adapter.DeriveAddress(seed, uint16(coinType), derivationPath, isDev)
	if err != nil {
		backendLogger.Error("derive address", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	// Returns publicKey and address as output
	return &logical.Response{
		Data: map[string]interface{}{
			"address":   address,
		},
	}, nil
}
