package tron

import (
	"encoding/hex"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/payment-system/dq-vault/lib/slip44"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSeed           = "test seed for deterministic key generation"
	testDerivationPath = "m/44'/195'/0'/0/0"
	validTronPath      = "0'/0/0"
	invalidPath        = "invalid/path"
	emptyPath          = ""
)

var (
	testSeedBytes = []byte(testSeed)
	logger        = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
)

func TestNewTronAdapter(t *testing.T) {
	tests := []struct {
		name             string
		logger           *slog.Logger
		expectedLogger   *slog.Logger
		expectedBasePath string
	}{
		{
			name:             "with valid logger",
			logger:           logger,
			expectedLogger:   logger,
			expectedBasePath: "44'/195'/",
		},
		{
			name:             "with nil logger",
			logger:           nil,
			expectedLogger:   nil,
			expectedBasePath: "44'/195'/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewTronAdapter(tt.logger)

			assert.NotNil(t, adapter)
			assert.Equal(t, tt.expectedLogger, adapter.logger)
			assert.Equal(t, tt.expectedBasePath, adapter.basePath)
		})
	}
}

func TestTronAdapter_CanDo(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name     string
		coinType uint16
		expected bool
	}{
		{
			name:     "tron supported",
			coinType: slip44.Tron,
			expected: true,
		},
		{
			name:     "bitcoin not supported",
			coinType: slip44.Bitcoin,
			expected: false,
		},
		{
			name:     "ethereum not supported",
			coinType: slip44.Ether,
			expected: false,
		},
		{
			name:     "litecoin not supported",
			coinType: slip44.Litecoin,
			expected: false,
		},
		{
			name:     "unknown coin type",
			coinType: 9999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.CanDo(tt.coinType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTronAdapter_parseDerivationPath(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name          string
		path          string
		expected      string
		expectedError error
	}{
		{
			name:          "valid relative path",
			path:          "m/0'/0/0",
			expected:      "44'/195'/0'/0/0",
			expectedError: nil,
		},
		{
			name:          "valid full tron path",
			path:          "m/44'/195'/0'/0/0",
			expected:      "44'/195'/0'/0/0",
			expectedError: nil,
		},
		{
			name:          "invalid full path (not tron)",
			path:          "m/44'/60'/0'/0/0",
			expected:      "",
			expectedError: ErrInvalidDerivationPath,
		},
		{
			name:          "empty path",
			path:          "",
			expected:      "",
			expectedError: ErrInvalidDerivationPath,
		},
		{
			name:          "invalid format",
			path:          "invalid/path",
			expected:      "",
			expectedError: ErrInvalidDerivationPath,
		},
		{
			name:          "path without m prefix",
			path:          "44'/195'/0'/0/0",
			expected:      "",
			expectedError: ErrInvalidDerivationPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.parseDerivationPath(tt.path)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestTronAdapter_DerivePrivateKey(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		expectError    bool
		expectedError  error
	}{
		{
			name:           "valid derivation",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "valid derivation with dev flag",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          true,
			expectError:    false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeedBytes,
			derivationPath: invalidPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty derivation path",
			seed:           testSeedBytes,
			derivationPath: emptyPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "nil seed",
			seed:           nil,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.DerivePrivateKey(tt.seed, tt.derivationPath, tt.isDev)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Verify it's a valid hex string
				_, hexErr := hex.DecodeString(result)
				assert.NoError(t, hexErr, "private key should be valid hex")
				// Private key should be 64 characters (32 bytes in hex)
				assert.Len(t, result, 64, "private key should be 64 hex characters")
			}
		})
	}
}

func TestTronAdapter_DerivePublicKey(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		expectError    bool
		expectedError  error
	}{
		{
			name:           "valid derivation",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "valid derivation with dev flag",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          true,
			expectError:    false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeedBytes,
			derivationPath: invalidPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty derivation path",
			seed:           testSeedBytes,
			derivationPath: emptyPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "nil seed",
			seed:           nil,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.DerivePublicKey(tt.seed, tt.derivationPath, tt.isDev)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Verify it's a valid hex string
				_, hexErr := hex.DecodeString(result)
				assert.NoError(t, hexErr, "public key should be valid hex")
				// Public key should be 130 characters (65 bytes in hex - uncompressed format with 04 prefix)
				assert.Len(t, result, 130, "public key should be 130 hex characters")
			}
		})
	}
}

func TestTronAdapter_DeriveAddress(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		isDev          bool
		expectError    bool
		expectedError  error
	}{
		{
			name:           "valid derivation",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "valid derivation with dev flag",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			isDev:          true,
			expectError:    false,
		},
		{
			name:           "invalid derivation path",
			seed:           testSeedBytes,
			derivationPath: invalidPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty derivation path",
			seed:           testSeedBytes,
			derivationPath: emptyPath,
			isDev:          false,
			expectError:    true,
			expectedError:  ErrInvalidDerivationPath,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
		{
			name:           "nil seed",
			seed:           nil,
			derivationPath: testDerivationPath,
			isDev:          false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.DeriveAddress(tt.seed, tt.derivationPath, tt.isDev)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Tron addresses should start with 'T'
				assert.True(t, strings.HasPrefix(result, "T"), "Tron address should start with 'T'")
				// Tron addresses should be 34 characters long
				assert.Len(t, result, 34, "Tron address should be 34 characters long")
			}
		})
	}
}

func TestTronAdapter_CreateSignedTransaction(t *testing.T) {
	adapter := NewTronAdapter(logger)

	// Create a minimal valid transaction raw data for testing
	// This is a simplified hex representation of a TransferContract
	validTransferPayload := "0a0208" // minimal valid protobuf for testing
	invalidPayload := "invalid_hex"
	emptyPayload := ""
	shortPayload := "0a"

	tests := []struct {
		name           string
		seed           []byte
		derivationPath string
		payload        string
		expectError    bool
		expectedError  error
	}{
		{
			name:           "invalid derivation path",
			seed:           testSeedBytes,
			derivationPath: invalidPath,
			payload:        validTransferPayload,
			expectError:    true,
			// Note: The error might be ErrInvalidRawData if path parsing happens later
		},
		{
			name:           "invalid hex payload",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			payload:        invalidPayload,
			expectError:    true,
		},
		{
			name:           "empty payload",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			payload:        emptyPayload,
			expectError:    true,
		},
		{
			name:           "short payload",
			seed:           testSeedBytes,
			derivationPath: testDerivationPath,
			payload:        shortPayload,
			expectError:    true,
		},
		{
			name:           "empty seed",
			seed:           []byte{},
			derivationPath: testDerivationPath,
			payload:        validTransferPayload,
			expectError:    true,
		},
		{
			name:           "nil seed",
			seed:           nil,
			derivationPath: testDerivationPath,
			payload:        validTransferPayload,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.CreateSignedTransaction(tt.seed, tt.derivationPath, tt.payload)

			if tt.expectError {
				assert.Error(t, err)
				// We don't check for specific error types here since the order of validation
				// can vary and multiple errors can occur
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				// Verify it's a valid hex string
				_, hexErr := hex.DecodeString(result)
				assert.NoError(t, hexErr, "signature should be valid hex")
			}
		})
	}
}

func TestTronAdapter_decodeContractData(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name          string
		contractData  string
		expectError   bool
		expectedError error
	}{
		{
			name:         "invalid hex data",
			contractData: "invalid_hex",
			expectError:  true,
		},
		{
			name:         "empty contract data",
			contractData: "",
			expectError:  true,
		},
		{
			name:         "too short contract data",
			contractData: "0x123",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := adapter.decodeContractData(tt.contractData)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestTronAdapter_trimFromFirstOfString(t *testing.T) {
	adapter := NewTronAdapter(logger)

	tests := []struct {
		name     string
		input    string
		n        int
		expected string
	}{
		{
			name:     "trim 2 characters",
			input:    "0x1234567890",
			n:        2,
			expected: "1234567890",
		},
		{
			name:     "trim 0 characters",
			input:    "1234567890",
			n:        0,
			expected: "1234567890",
		},
		{
			name:     "trim more than string length",
			input:    "123",
			n:        10,
			expected: "",
		},
		{
			name:     "trim from empty string",
			input:    "",
			n:        2,
			expected: "",
		},
		{
			name:     "trim exact string length",
			input:    "123",
			n:        3,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := adapter.trimFromFirstOfString(tt.input, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTronAdapter_Integration tests the integration between methods
func TestTronAdapter_Integration(t *testing.T) {
	adapter := NewTronAdapter(logger)

	// Test that derived keys are consistent
	t.Run("derived keys consistency", func(t *testing.T) {
		privateKey1, err1 := adapter.DerivePrivateKey(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err1)

		privateKey2, err2 := adapter.DerivePrivateKey(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err2)

		// Same seed and path should produce same private key
		assert.Equal(t, privateKey1, privateKey2)

		publicKey1, err3 := adapter.DerivePublicKey(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err3)

		publicKey2, err4 := adapter.DerivePublicKey(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err4)

		// Same seed and path should produce same public key
		assert.Equal(t, publicKey1, publicKey2)

		address1, err5 := adapter.DeriveAddress(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err5)

		address2, err6 := adapter.DeriveAddress(testSeedBytes, testDerivationPath, false)
		require.NoError(t, err6)

		// Same seed and path should produce same address
		assert.Equal(t, address1, address2)
	})

	// Test that different seeds produce different results
	t.Run("different seeds produce different results", func(t *testing.T) {
		seed1 := []byte("seed1")
		seed2 := []byte("seed2")

		privateKey1, err1 := adapter.DerivePrivateKey(seed1, testDerivationPath, false)
		require.NoError(t, err1)

		privateKey2, err2 := adapter.DerivePrivateKey(seed2, testDerivationPath, false)
		require.NoError(t, err2)

		assert.NotEqual(t, privateKey1, privateKey2)

		address1, err3 := adapter.DeriveAddress(seed1, testDerivationPath, false)
		require.NoError(t, err3)

		address2, err4 := adapter.DeriveAddress(seed2, testDerivationPath, false)
		require.NoError(t, err4)

		assert.NotEqual(t, address1, address2)
	})

	// Test that different paths produce different results
	t.Run("different paths produce different results", func(t *testing.T) {
		path1 := "m/44'/195'/0'/0/0"
		path2 := "m/44'/195'/0'/0/1"

		privateKey1, err1 := adapter.DerivePrivateKey(testSeedBytes, path1, false)
		require.NoError(t, err1)

		privateKey2, err2 := adapter.DerivePrivateKey(testSeedBytes, path2, false)
		require.NoError(t, err2)

		assert.NotEqual(t, privateKey1, privateKey2)

		address1, err3 := adapter.DeriveAddress(testSeedBytes, path1, false)
		require.NoError(t, err3)

		address2, err4 := adapter.DeriveAddress(testSeedBytes, path2, false)
		require.NoError(t, err4)

		assert.NotEqual(t, address1, address2)
	})
}
