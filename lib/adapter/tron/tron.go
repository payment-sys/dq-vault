package tron

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fbsobreira/gotron-sdk/pkg/address"
	"github.com/fbsobreira/gotron-sdk/pkg/common"
	"github.com/fbsobreira/gotron-sdk/pkg/keys/hd"
	"github.com/fbsobreira/gotron-sdk/pkg/proto/core"
	"github.com/payment-system/dq-vault/lib/slip44"
	"google.golang.org/protobuf/proto"
)

const (
	// maskingLength is the number of characters to show at the end of masked keys
	maskingLength          = 4
	tronHexAddressPrefix   = "41"
	relativePathComponents = 3
	hexPrefixLength        = 2
	methodSignatureLength  = 10
)

// Adapter represents a Tron blockchain adapter
type Adapter struct {
	logger   *slog.Logger
	basePath string
}

// NewTronAdapter creates a new Tron adapter instance
func NewTronAdapter(logger *slog.Logger) *Adapter {
	return &Adapter{
		logger:   logger,
		basePath: "44'/195'/",
	}
}

// CanDo checks if this adapter can handle the given coin type
func (t *Adapter) CanDo(coinType uint16) bool {
	return coinType == slip44.Tron
}

func (t *Adapter) parseDerivationPath(path string) (string, error) {
	logger := t.logger.With(slog.String("op", "parse_derivation_path"), slog.String("path", path))
	logger.Info("Parsing derivation path")

	components := strings.Split(path, "/")
	switch {
	case strings.TrimSpace(components[0]) == "":
		return "", ErrInvalidDerivationPath

	case strings.TrimSpace(components[0]) == "m":
		components = components[1:]
		subParts := strings.Join(components, "/")

		// if path is relative, append tron base path to derivation path
		if len(components) == relativePathComponents {
			return t.basePath + subParts, nil

			// if it's full path but it's not tron (195)
		} else if !strings.Contains(subParts, t.basePath) {
			return "", ErrInvalidDerivationPath
		}

		return subParts, nil
	default:
		return "", ErrInvalidDerivationPath
	}
}

func (t *Adapter) deriveKeysForPath(seed []byte, derivationPath string) (
	*secp256k1.PrivateKey, *secp256k1.PublicKey, error) {
	derivationPath, err := t.parseDerivationPath(derivationPath)
	if err != nil {
		return nil, nil, err
	}

	master, ch := hd.ComputeMastersFromSeed(seed, []byte("Bitcoin seed"))
	private, err := hd.DerivePrivateKeyForPath(
		btcec.S256(),
		master,
		ch,
		derivationPath,
	)
	if err != nil {
		return nil, nil, err
	}

	privateKey, publicKey := btcec.PrivKeyFromBytes(private[:])
	return privateKey, publicKey, nil
}

// DerivePrivateKey derives a private key from the given seed and derivation path
func (t *Adapter) DerivePrivateKey(seed []byte, derivationPath string, _ bool) (string, error) {
	logger := t.logger.With(slog.String("op", "derive_private_key"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving private key")

	privateKey, _, err := t.deriveKeysForPath(seed, derivationPath)
	if err != nil {
		logger.Error("Failed to derive private key", "error", err)
		return "", err
	}

	privateKeyBytes := crypto.FromECDSA(privateKey.ToECDSA())

	// bytes to hex encoded string
	// excluding "0x" prefix
	privateKeyHex := hexutil.Encode(privateKeyBytes)[hexPrefixLength:]

	// mask the private key
	maskedPrivateKey := privateKeyHex[:len(privateKeyHex)-maskingLength] + strings.Repeat("*", maskingLength)
	logger.Info("Private key", "privateKey", maskedPrivateKey)

	return privateKeyHex, nil
}

// DerivePublicKey derives a public key from the given seed and derivation path
func (t *Adapter) DerivePublicKey(seed []byte, derivationPath string, _ bool) (string, error) {
	logger := t.logger.With(slog.String("op", "derive_public_key"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving public key")

	_, publicKey, err := t.deriveKeysForPath(seed, derivationPath)
	if err != nil {
		logger.Error("Failed to derive public key", "error", err)
		return "", err
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKey.ToECDSA())
	publicKeyHex := hexutil.Encode(publicKeyBytes)[hexPrefixLength:]

	// mask the public key
	maskedPublicKey := publicKeyHex[:len(publicKeyHex)-maskingLength] + strings.Repeat("*", maskingLength)
	logger.Info("Public key", "publicKey", maskedPublicKey)

	return publicKeyHex, nil
}

// DeriveAddress derives an address from the given seed and derivation path
func (t *Adapter) DeriveAddress(seed []byte, derivationPath string, _ bool) (string, error) {
	logger := t.logger.With(slog.String("op", "derive_address"), slog.String("derivationPath", derivationPath))
	logger.Info("Deriving address")

	_, publicKey, err := t.deriveKeysForPath(seed, derivationPath)
	if err != nil {
		logger.Error("Failed to derive address", "error", err)
		return "", err
	}

	tronAddress := address.PubkeyToAddress(*publicKey.ToECDSA())

	logger.Info("Address", "address", tronAddress.String())

	return tronAddress.String(), nil
}

// CreateSignedTransaction creates a signed transaction from the given parameters
func (t *Adapter) CreateSignedTransaction(seed []byte, derivationPath, payload string) (string, error) {
	logger := t.logger.With(slog.String("op", "create_signed_transaction"), slog.String("derivationPath", derivationPath))
	logger.Info("Creating signed transaction")

	raw := &core.TransactionRaw{}
	decodedHex, err := hex.DecodeString(payload)
	if err != nil {
		return "", err
	}

	err = proto.Unmarshal(decodedHex, raw)
	if err != nil {
		return "", ErrInvalidRawData
	}

	contracts := raw.GetContract()
	if len(contracts) == 0 {
		return "", ErrInvalidRawData
	}

	c := contracts[0]
	contractType := c.GetType()

	var toAddress string

	switch contractType {
	case core.Transaction_Contract_TransferContract:
		transfer := &core.TransferContract{}
		err = c.GetParameter().UnmarshalTo(transfer)
		if err != nil {
			return "", err
		}

		toAddress = common.EncodeCheck(transfer.ToAddress)

	case core.Transaction_Contract_TriggerSmartContract:
		trigger := &core.TriggerSmartContract{}
		err = c.GetParameter().UnmarshalTo(trigger)
		if err != nil {
			return "", err
		}

		txContractData := "0x" + hex.EncodeToString(trigger.GetData())

		toAddressInHexFormat, err := t.decodeContractData(txContractData)
		if err != nil {
			return "", err
		}

		decodedToTronHexWallet, err := hex.DecodeString(toAddressInHexFormat)
		if err != nil {
			return "", err
		}

		toAddress = common.EncodeCheck(decodedToTronHexWallet)

	default:
		return "", ErrUnsupportedTransactionType
	}

	logger.Info("To address", "toAddress", toAddress)

	privateKey, err := t.DerivePrivateKey(seed, derivationPath, false)
	if err != nil {
		return "", err
	}

	privateBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return "", err
	}

	ecdsaPrivateKey, err := crypto.ToECDSA(privateBytes)
	if err != nil {
		return "", err
	}

	rawDataJSON, err := proto.Marshal(raw)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(rawDataJSON)
	txIDHash := h.Sum(nil)

	sig, err := crypto.Sign(txIDHash, ecdsaPrivateKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(sig), nil
}

func (t *Adapter) decodeContractData(txContractData string) (string, error) {
	// Check if txContractData is long enough
	if len(txContractData) < methodSignatureLength {
		return "", ErrInvalidRawData
	}

	// load contract ABI
	abi2, err := abi.JSON(strings.NewReader(TRC20ABI))
	if err != nil {
		return "", err
	}

	// decode txInput method signature
	decodedSig, err := hex.DecodeString(txContractData[hexPrefixLength:methodSignatureLength])
	if err != nil {
		return "", err
	}

	// recover Method from signature and ABI
	method, err := abi2.MethodById(decodedSig)
	if err != nil {
		return "", err
	}

	// Check if txContractData is long enough for payload
	if len(txContractData) <= methodSignatureLength {
		return "", ErrInvalidRawData
	}

	// decode txInput Payload
	decodedData, err := hex.DecodeString(txContractData[methodSignatureLength:])
	if err != nil {
		return "", err
	}

	// unpack method inputs
	data, err := method.Inputs.Unpack(decodedData)
	if err != nil {
		return "", err
	}

	var toTronHexWallet, to string

	/* we do this because of method calling */
	if len(data) == 2 || method.Name == "transfer" {
		/* Method Calling = transfer */
		to = fmt.Sprintf("%v", data[0])
	} else if len(data) == 3 || method.Name == "transferFrom" {
		/* Method Calling = transferFrom */
		to = fmt.Sprintf("%v", data[1])
	}

	// to eliminate 0x
	walletWithOut0x := t.trimFromFirstOfString(to, hexPrefixLength)
	toTronHexWallet = tronHexAddressPrefix + walletWithOut0x

	return toTronHexWallet, nil
}

func (*Adapter) trimFromFirstOfString(s string, n int) string {
	m := 0
	for i := range s {
		if m >= n {
			return s[i:]
		}
		m++
	}
	return s[:0]
}
