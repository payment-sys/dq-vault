package api

import (
	"context"
	"fmt"
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

// pathAddressBatch generates a batch of addresses using a templated derivation path.
func (b *Backend) pathAddressBatch(ctx context.Context, req *logical.Request,
	d *framework.FieldData) (*logical.Response, error) {
	backendLogger := b.logger.With(slog.String("op", "path_address_batch"))
	if err := helpers.ValidateFields(req, d); err != nil {
		backendLogger.Error("validate fields", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	uuid := d.Get("uuid").(string)
	pathTemplate := d.Get("pathTemplate").(string)
	coinType := d.Get("coinType").(int)
	isDev := d.Get("isDev").(bool)
	startIndex := d.Get("startIndex").(int)
	count := d.Get("count").(int)

	if count <= 0 || count > 1000 {
		return nil, logical.CodedError(http.StatusBadRequest, "count must be between 1 and 1000")
	}

	if uint16(coinType) == slip44.Bitshares {
		pathTemplate = config.BitsharesDerivationPath
	}

	// Validate base data
	if err := helpers.ValidateData(ctx, req, uuid, pathTemplate); err != nil {
		backendLogger.Error("validate data", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	// Retrieve user info
	path := config.StorageBasePath + uuid
	entry, err := req.Storage.Get(ctx, path)
	if err != nil {
		backendLogger.Error("get", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	var userInfo helpers.User
	err = entry.DecodeJSON(&userInfo)
	if err != nil {
		backendLogger.Error("decode json", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	seed, err := lib.SeedFromMnemonic(userInfo.Mnemonic, userInfo.Passphrase)
	if err != nil {
		backendLogger.Error("seed from mnemonic", "error", err)
		return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
	}

	adapterInventory := adapter.GetInventory(backendLogger)

	addresses := make(map[string]string, count)
	for i := startIndex; i < startIndex+count; i++ {
		var derivationPath string
		if pathTemplate == config.BitsharesDerivationPath {
			derivationPath = pathTemplate
		} else {
			derivationPath = fmt.Sprintf(pathTemplate, i)
		}
		address, err := adapterInventory.DeriveAddress(seed, uint16(coinType), derivationPath, isDev)
		if err != nil {
			backendLogger.Error("derive address", "error", err, "index", i)
			return nil, logical.CodedError(http.StatusUnprocessableEntity, err.Error())
		}
		addresses[derivationPath] = address
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"addresses": addresses,
		},
	}, nil
}
