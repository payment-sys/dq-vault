package evm

import (
	"encoding/hex"
	"log/slog"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/payment-system/dq-vault/lib"
	"github.com/payment-system/dq-vault/lib/slip44"
)

// Test constants
const (
	// Test seed derived from mnemonic
	testSeedHex = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
	// Test derivation path
	testDerivationPath = "m/44'/60'/0'/0/0"
	// Expected private key for test seed + path
	expectedPrivateKey = "1ab42cc412b618bdea3a599e3c9bae199ebf030895b039e9db1e30dafb12b727"
	// Expected public key for test seed + path
	expectedPublicKey = "0237b0bb7a8288d38ed49a524b5dc98cff3eb5ca824c9f9dc0dfdb3d9cd600f299"
	// Expected address for test seed + path
	expectedAddress = "0x9858EfFD232B4033E47d90003D41EC34EcaEda94"
)

func TestNewEthereumAdapter(t *testing.T) {
	tests := []struct {
		name   string
		logger *slog.Logger
		want   *EthereumAdapter
	}{
		{
			name:   "with valid logger",
			logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
			want: &EthereumAdapter{
				availableCoinTypes: []uint16{
					slip44.Ether,
					slip44.Binance,
					slip44.Polygon,
					slip44.Avalanche,
					slip44.Fantom,
					slip44.Harmony,
				},
				zeroAddress: "0x0000000000000000000000000000000000000000",
			},
		},
		{
			name:   "with text handler logger",
			logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
			want: &EthereumAdapter{
				availableCoinTypes: []uint16{
					slip44.Ether,
					slip44.Binance,
					slip44.Polygon,
					slip44.Avalanche,
					slip44.Fantom,
					slip44.Harmony,
				},
				zeroAddress: "0x0000000000000000000000000000000000000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewEthereumAdapter(tt.logger)
			assert.NotNil(t, got)
			assert.Equal(t, tt.want.availableCoinTypes, got.availableCoinTypes)
			assert.Equal(t, tt.want.zeroAddress, got.zeroAddress)
			assert.NotNil(t, got.logger)
		})
	}
}

func TestEthereumAdapter_CanDo(t *testing.T) {
	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tests := []struct {
		name     string
		coinType uint16
		want     bool
	}{
		{
			name:     "ethereum supported",
			coinType: slip44.Ether,
			want:     true,
		},
		{
			name:     "binance supported",
			coinType: slip44.Binance,
			want:     true,
		},
		{
			name:     "polygon supported",
			coinType: slip44.Polygon,
			want:     true,
		},
		{
			name:     "avalanche supported",
			coinType: slip44.Avalanche,
			want:     true,
		},
		{
			name:     "fantom supported",
			coinType: slip44.Fantom,
			want:     true,
		},
		{
			name:     "harmony supported",
			coinType: slip44.Harmony,
			want:     true,
		},
		{
			name:     "bitcoin not supported",
			coinType: slip44.Bitcoin,
			want:     false,
		},
		{
			name:     "litecoin not supported",
			coinType: slip44.Litecoin,
			want:     false,
		},
		{
			name:     "monero not supported",
			coinType: slip44.Monero,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := adapter.CanDo(tt.coinType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEthereumAdapter_DerivePrivateKey(t *testing.T) {
	testSeed, err := hex.DecodeString(testSeedHex)
	require.NoError(t, err)

	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		want           string
		wantErr        bool
	}{
		{
			name:           "valid mainnet derivation",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           expectedPrivateKey,
			wantErr:        false,
		},
		{
			name:           "valid testnet derivation",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			isDev:          true,
			want:           expectedPrivateKey,
			wantErr:        false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeed,
			derivationPath: "invalid/path",
			isDev:          false,
			want:           "",
			wantErr:        true,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           "5e9340935f4c02628cec5d04cc281012537cafa8dae0e27ff56563b8dffab368",
			wantErr:        false,
		},
		{
			name:           "nil seed",
			seed:           nil,
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           "5e9340935f4c02628cec5d04cc281012537cafa8dae0e27ff56563b8dffab368",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adapter.DerivePrivateKey(tt.seed, tt.derivationPath, tt.isDev)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestEthereumAdapter_DerivePublicKey(t *testing.T) {
	testSeed, err := hex.DecodeString(testSeedHex)
	require.NoError(t, err)

	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		want           string
		wantErr        bool
	}{
		{
			name:           "valid derivation",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           expectedPublicKey,
			wantErr:        false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeed,
			derivationPath: "invalid/path",
			isDev:          false,
			want:           "",
			wantErr:        true,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           "038840222aff65bf7bc40ac0b832935f6b3967af0fc7b37feca31300f03344e3ce",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adapter.DerivePublicKey(tt.seed, tt.derivationPath, tt.isDev)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.NotEmpty(t, got)
			}
		})
	}
}

func TestEthereumAdapter_DeriveAddress(t *testing.T) {
	testSeed, err := hex.DecodeString(testSeedHex)
	require.NoError(t, err)

	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		want           string
		wantErr        bool
	}{
		{
			name:           "valid derivation",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           expectedAddress,
			wantErr:        false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeed,
			derivationPath: "invalid/path",
			isDev:          false,
			want:           "",
			wantErr:        true,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			want:           "0x1667CA2C72D8699f0C34c55ea00b60Eef021Be3a",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adapter.DeriveAddress(tt.seed, tt.derivationPath, tt.isDev)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
				assert.NotEmpty(t, got)
				assert.True(t, len(got) == 42) // Ethereum addresses are 42 characters
				if len(got) >= 2 {
					assert.True(t, got[:2] == "0x") // Ethereum addresses start with 0x
				}
			}
		})
	}
}

func TestEthereumAdapter_CreateSignedTransaction(t *testing.T) {
	testSeed, err := hex.DecodeString(testSeedHex)
	require.NoError(t, err)

	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	// Valid Ethereum transaction payload
	validPayload := `{
		"nonce": 42,
		"value": 1000000000000000000,
		"gasLimit": 21000,
		"gasPrice": 20000000000,
		"to": "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
		"data": "0x",
		"chainId": 1
	}`

	// Invalid payload - zero address which is considered invalid
	invalidPayload := `{
		"nonce": 42,
		"value": 1000000000000000000,
		"gasLimit": 21000,
		"gasPrice": 20000000000,
		"to": "0x0000000000000000000000000000000000000000",
		"data": "0x",
		"chainId": 1
	}`

	// Invalid payload - malformed JSON
	malformedPayload := `{invalid json`

	// Invalid payload - negative values
	negativeValuePayload := `{
		"nonce": 42,
		"value": -1000000000000000000,
		"gasLimit": 21000,
		"gasPrice": 20000000000,
		"to": "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
		"data": "0x",
		"chainId": 1
	}`

	// Invalid payload - invalid address
	invalidAddressPayload := `{
		"nonce": 42,
		"value": 1000000000000000000,
		"gasLimit": 21000,
		"gasPrice": 20000000000,
		"to": "invalid-address",
		"data": "0x",
		"chainId": 1
	}`

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		payload        string
		wantErr        bool
		wantErrType    error
	}{
		{
			name:           "valid transaction",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			payload:        validPayload,
			wantErr:        false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeed,
			derivationPath: "invalid/path",
			payload:        validPayload,
			wantErr:        true,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			payload:        validPayload,
			wantErr:        false,
		},
		{
			name:           "invalid payload - zero address",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			payload:        invalidPayload,
			wantErr:        true,
		},
		{
			name:           "malformed JSON payload",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			payload:        malformedPayload,
			wantErr:        true,
		},
		{
			name:           "negative value payload",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			payload:        negativeValuePayload,
			wantErr:        true,
		},
		{
			name:           "invalid address payload",
			seed:           testSeed,
			derivationPath: testDerivationPath,
			payload:        invalidAddressPayload,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := adapter.CreateSignedTransaction(tt.seed, tt.derivationPath, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, got)
				assert.True(t, len(got) > 0)
				if len(got) >= 2 {
					assert.True(t, got[:2] == "0x") // Signed transactions start with 0x
				}
			}
		})
	}
}

func TestValidatePayload(t *testing.T) {
	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	tests := []struct {
		name    string
		payload lib.EthereumRawTx
		want    bool
		wantTx  string
	}{
		{
			name: "valid ether transfer",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
				To:       "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
				Data:     "",
				ChainID:  big.NewInt(1),
			},
			want:   true,
			wantTx: "Ether Transfer",
		},
		{
			name: "valid contract function call",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(0),
				GasLimit: 50000,
				GasPrice: big.NewInt(20000000000),
				To:       "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
				Data:     "0xa9059cbb000000000000000000000000742d35cc6634c0532925a3b8d359a5c5119e32c8",
				ChainID:  big.NewInt(1),
			},
			want:   true,
			wantTx: "Contract Function Call",
		},
		{
			name: "valid contract creation",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(0),
				GasLimit: 500000,
				GasPrice: big.NewInt(20000000000),
				To:       "",
				Data:     "0x608060405234801561001057600080fd5b50",
				ChainID:  big.NewInt(1),
			},
			want:   true,
			wantTx: "Contract Creation",
		},
		{
			name: "invalid negative value",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(-1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
				To:       "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
				Data:     "0x",
				ChainID:  big.NewInt(1),
			},
			want:   false,
			wantTx: "",
		},
		{
			name: "invalid negative gas price",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(-20000000000),
				To:       "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
				Data:     "0x",
				ChainID:  big.NewInt(1),
			},
			want:   false,
			wantTx: "",
		},
		{
			name: "invalid negative chain id",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
				To:       "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
				Data:     "0x",
				ChainID:  big.NewInt(-1),
			},
			want:   false,
			wantTx: "",
		},
		{
			name: "invalid address format",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
				To:       "invalid-address",
				Data:     "0x",
				ChainID:  big.NewInt(1),
			},
			want:   false,
			wantTx: "",
		},
		{
			name: "zero address",
			payload: lib.EthereumRawTx{
				Nonce:    42,
				Value:    big.NewInt(1000000000000000000),
				GasLimit: 21000,
				GasPrice: big.NewInt(20000000000),
				To:       "0x0000000000000000000000000000000000000000",
				Data:     "0x",
				ChainID:  big.NewInt(1),
			},
			want:   false,
			wantTx: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotTx := validatePayload(tt.payload, adapter.zeroAddress)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantTx, gotTx)
		})
	}
}

// Benchmark tests
func BenchmarkEthereumAdapter_DerivePrivateKey(b *testing.B) {
	testSeed, _ := hex.DecodeString(testSeedHex)
	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.DerivePrivateKey(testSeed, testDerivationPath, false)
	}
}

func BenchmarkEthereumAdapter_DeriveAddress(b *testing.B) {
	testSeed, _ := hex.DecodeString(testSeedHex)
	adapter := NewEthereumAdapter(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = adapter.DeriveAddress(testSeed, testDerivationPath, false)
	}
}
