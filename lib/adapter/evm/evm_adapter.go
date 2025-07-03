package evm

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"reflect"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/payment-system/dq-vault/lib"
	"github.com/payment-system/dq-vault/lib/slip44"
)


type EthereumAdapter struct {
	logger *slog.Logger
	availableCoinTypes []uint16
	zeroAddress string
}

func NewEthereumAdapter(logger *slog.Logger) *EthereumAdapter {
	return &EthereumAdapter{
		logger: logger.With(slog.String("adapter", "evm")),
		availableCoinTypes: []uint16{
			slip44.Ether,
			slip44.Binance,
			slip44.Polygon,
			slip44.Avalanche,
			slip44.Fantom,
			slip44.Harmony,
		},
		zeroAddress: "0x0000000000000000000000000000000000000000",
	}
}

func (e *EthereumAdapter) CanDo(coinType uint16) bool {
	return slices.Contains(e.availableCoinTypes, coinType)
}

func (e *EthereumAdapter) DerivePrivateKey(seed []byte, derivationPath string, isDev bool) (string, error) {
	logger := e.logger.With(slog.String("op", "derive_private_key"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving private key")

	btcecPrivateKey, err := lib.DerivePrivateKey(seed, derivationPath, isDev)
	if err != nil {
		logger.Error("Failed to derive private key", "error", err)
		return "", err
	}

	privateKey := crypto.FromECDSA(btcecPrivateKey.ToECDSA())
	privateKeyStr := hexutil.Encode(privateKey)[2:]

	maskedKey := strings.Repeat("*", len(privateKeyStr)-4) + privateKeyStr[len(privateKeyStr)-4:]
	logger.Info("Private key derived successfully", "privateKey", maskedKey)

	return privateKeyStr, nil
}

func (e *EthereumAdapter) DerivePublicKey(seed []byte, derivationPath string, isDev bool) (string, error) {
	logger := e.logger.With(slog.String("op", "derive_public_key"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving public key")

	prvKey, err := e.DerivePrivateKey(seed, derivationPath, isDev)
	if err != nil {
		logger.Error("Failed to derive private key", "error", err)
		return "", err
	}
	privateKey, err := crypto.HexToECDSA(prvKey)
	if err != nil {
		return "", err
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return "", ErrInvalidECDSAPublicKey
	}

	publicKeyBytes := crypto.CompressPubkey(publicKeyECDSA)
	publicKeyStr := hexutil.Encode(publicKeyBytes)[2:]

	maskedKey := strings.Repeat("*", len(publicKeyStr)-4) + publicKeyStr[len(publicKeyStr)-4:]
	logger.Info("Public key derived successfully", "publicKey", maskedKey)

	return publicKeyStr, nil
}

func (e *EthereumAdapter) DeriveAddress(seed []byte, derivationPath string, isDev bool) (string, error) {
	logger := e.logger.With(slog.String("op", "derive_address"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving address")

	prvKey, err := e.DerivePrivateKey(seed, derivationPath, isDev)
	if err != nil {
		logger.Error("Failed to derive private key", "error", err)
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(prvKey)
	if err != nil {
		return "", err
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return "", ErrInvalidECDSAPublicKey
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	logger.Info("Address derived successfully", "address", address)

	return address, nil
}

func validatePayload(payload lib.EthereumRawTx, zeroAddress string) (bool, string) {
	// Value, chainId, GasPrice should not be negative
	if payload.ChainID.Cmp(big.NewInt(0)) == -1 ||
		payload.Value.Cmp(big.NewInt(0)) == -1 ||
		payload.GasPrice.Cmp(big.NewInt(0)) == -1 {
		return false, ""
	}

	if payload.To == "" && payload.Data != "" {
		return true, "Contract Creation"
	}

	if payload.To != "" {
		if !common.IsHexAddress(payload.To) ||
			!strings.HasPrefix(payload.To, "0x") || len(payload.To) != 42 ||
			payload.To == zeroAddress {
			return false, ""
		}
		transactionType := "Ether Transfer"
		if payload.Data != "" {
			transactionType = "Contract Function Call"
		}

		return true, transactionType
	}
	return false, ""
}

func (e *EthereumAdapter) createRawTransaction(payloadString string) (*types.Transaction, *big.Int, error) {
	logger := e.logger.With(slog.String("op", "create_raw_transaction"))
	logger.Info("Creating raw transaction")

	var payload lib.EthereumRawTx
	if err := json.Unmarshal([]byte(payloadString), &payload); err != nil ||
		reflect.DeepEqual(payload, lib.EthereumRawTx{}) {
		errorMsg := fmt.Sprintf("Unable to decode payload=[%v]", payloadString)

		return nil, nil, errors.New(errorMsg)
	}

	// validate payload data
	valid, txType := validatePayload(payload, e.zeroAddress)
	if !valid {
		return nil, nil, errors.New("Invalid payload data")
	}

	logger.Info("validate payload", "txType", txType)
	// create raw transaction from payload data
	return types.NewTransaction(
		payload.Nonce,
		common.HexToAddress(payload.To),
		payload.Value,
		payload.GasLimit,
		payload.GasPrice,
		common.FromHex(string(payload.Data)),
	), payload.ChainID, nil
}

func (e *EthereumAdapter) CreateSignedTransaction(seed []byte, derivationPath string, payload string) (string, error) {
	logger := e.logger.With(slog.String("op", "create_signed_transaction"), slog.String("derivationPath", derivationPath))
	logger.Info("Creating signed transaction")

	prvKey, err := e.DerivePrivateKey(seed, derivationPath, false)
	if err != nil {
		logger.Error("Failed to derive private key", "error", err)
		return "", err
	}

	privateKey, err := crypto.HexToECDSA(prvKey)
	if err != nil {
		return "", err
	}

	rawTx, chainId, err := e.createRawTransaction(payload)
	if err != nil {
		logger.Error("Failed to create raw transaction", "error", err)
		return "", err
	}

	// sign raw transaction using raw transaction + chainId + private key
	signedTx, err := types.SignTx(rawTx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		return "", err
	}
	// obtains signed transaction hex
	var signedTxBuff bytes.Buffer
	_ = signedTx.EncodeRLP(&signedTxBuff)
	txHex := hexutil.Encode(signedTxBuff.Bytes())

	logger.Info("Signed transaction created successfully", "tx", txHex)

	return txHex, nil
}