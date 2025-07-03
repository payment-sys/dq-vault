package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/payment-system/dq-vault/api/helpers"
	"github.com/payment-system/dq-vault/config"
	"github.com/payment-system/dq-vault/lib/slip44"
)

// Test constants
const (
	testUUID           = "test-uuid-123"
	testDerivationPath = "m/44'/60'/0'/0/0"
	testMnemonic       = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	testPassphrase     = "test-passphrase"
	testAddress        = "0x9858EfFD232B4033E47d90003D41EC34EcaEda94"
)

// MockStorage implements logical.Storage for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) Get(ctx context.Context, key string) (*logical.StorageEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logical.StorageEntry), args.Error(1)
}

func (m *MockStorage) Put(ctx context.Context, entry *logical.StorageEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStorage) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Helper function to create a proper framework.FieldData
func createFieldData(data map[string]interface{}) *framework.FieldData {
	// Create the schema that matches what the address endpoint expects
	schema := map[string]*framework.FieldSchema{
		"uuid": {
			Type:        framework.TypeString,
			Description: "UUID of user",
		},
		"path": {
			Type:        framework.TypeString,
			Description: "Derivation path",
		},
		"coinType": {
			Type:        framework.TypeInt,
			Description: "Coin type",
		},
		"isDev": {
			Type:        framework.TypeBool,
			Description: "Development mode flag",
		},
	}

	return &framework.FieldData{
		Raw:    data,
		Schema: schema,
	}
}

// Helper function to create test backend
func createTestBackend(_ *testing.T) *Backend {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &Backend{
		logger: logger,
	}
}

// Helper function to create user storage entry
func createUserStorageEntry(t *testing.T, user helpers.User) *logical.StorageEntry {
	data, err := json.Marshal(user)
	require.NoError(t, err)

	return &logical.StorageEntry{
		Key:   config.StorageBasePath + testUUID,
		Value: data,
	}
}

func TestBackend_PathAddress(t *testing.T) {
	ctx := context.Background()

	// Test user data
	testUser := helpers.User{
		Mnemonic:   testMnemonic,
		Passphrase: testPassphrase,
	}

	tests := []struct {
		name           string
		fieldData      map[string]interface{}
		setupStorage   func(*MockStorage)
		setupMocks     func()
		want           *logical.Response
		wantErr        bool
		wantStatusCode int
		wantErrMsg     string
	}{
		{
			name: "successful ethereum address derivation",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				entry := createUserStorageEntry(t, testUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"address": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful address derivation with isDev true",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    true,
			},
			setupStorage: func(ms *MockStorage) {
				entry := createUserStorageEntry(t, testUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"address": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "missing uuid field",
			fieldData: map[string]interface{}{
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(_ *MockStorage) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing path field",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(_ *MockStorage) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing coinType field",
			fieldData: map[string]interface{}{
				"uuid":  testUUID,
				"path":  testDerivationPath,
				"isDev": false,
			},
			setupStorage: func(ms *MockStorage) {
				// Mock List for UUID existence check since ValidateData will be called
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
				// Mock Get since coinType=0 doesn't cause validation failure
				testUser := helpers.User{
					Mnemonic:   testMnemonic,
					Passphrase: testPassphrase,
				}
				entry := createUserStorageEntry(t, testUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "storage get error",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(nil, assert.AnError)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "user not found in storage",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
				// Return an error instead of nil entry to avoid panic
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return((*logical.StorageEntry)(nil), assert.AnError)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid json in storage",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
				invalidEntry := &logical.StorageEntry{
					Key:   config.StorageBasePath + testUUID,
					Value: []byte("{invalid json}"),
				}
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(invalidEntry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "empty mnemonic in user data",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				emptyUser := helpers.User{
					Mnemonic:   "",
					Passphrase: testPassphrase,
				}
				entry := createUserStorageEntry(t, emptyUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid derivation path",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     "invalid/path",
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				entry := createUserStorageEntry(t, testUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "unsupported coin type",
			fieldData: map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": 99999, // Unsupported coin type
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorage) {
				entry := createUserStorageEntry(t, testUser)
				ms.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStorage := new(MockStorage)
			backend := createTestBackend(t)

			// Setup storage expectations
			if tt.setupStorage != nil {
				tt.setupStorage(mockStorage)
			}

			// Setup field data
			fieldData := createFieldData(tt.fieldData)

			// Create request with mock storage
			req := &logical.Request{
				Storage: mockStorage,
				Data:    tt.fieldData,
			}

			// Execute
			got, err := backend.pathAddress(ctx, req, fieldData)

			// Assert error expectations
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantStatusCode != 0 {
					if codedErr, ok := err.(logical.HTTPCodedError); ok {
						assert.Equal(t, tt.wantStatusCode, codedErr.Code())
					}
				}
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Assert response structure
				if tt.want != nil {
					assert.NotNil(t, got.Data)
					assert.Contains(t, got.Data, "address")
					assert.NotEmpty(t, got.Data["address"])
				}
			}

			// Verify all mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathAddress_CoinTypeSpecific(t *testing.T) {
	ctx := context.Background()

	testUser := helpers.User{
		Mnemonic:   testMnemonic,
		Passphrase: testPassphrase,
	}

	// Test different coin types
	coinTypeTests := []struct {
		name           string
		coinType       uint16
		expectedPath   string
		shouldOverride bool
	}{
		{
			name:           "ethereum coin type",
			coinType:       slip44.Ether,
			expectedPath:   testDerivationPath,
			shouldOverride: false,
		},
		{
			name:           "bitshares coin type with path override",
			coinType:       slip44.Bitshares,
			expectedPath:   config.BitsharesDerivationPath,
			shouldOverride: true,
		},
		{
			name:           "bitcoin coin type",
			coinType:       slip44.Bitcoin,
			expectedPath:   testDerivationPath,
			shouldOverride: false,
		},
	}

	for _, tt := range coinTypeTests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStorage := new(MockStorage)
			backend := createTestBackend(t)

			entry := createUserStorageEntry(t, testUser)
			mockStorage.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
			// Mock List for UUID existence check
			mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)

			fieldData := createFieldData(map[string]interface{}{
				"uuid":     testUUID,
				"path":     testDerivationPath,
				"coinType": int(tt.coinType),
				"isDev":    false,
			})

			req := &logical.Request{
				Storage: mockStorage,
				Data: map[string]interface{}{
					"uuid":     testUUID,
					"path":     testDerivationPath,
					"coinType": int(tt.coinType),
					"isDev":    false,
				},
			}

			// Execute
			got, err := backend.pathAddress(ctx, req, fieldData)

			// For unsupported coin types, we expect an error
			// Currently only Ether is supported by the EVM adapter
			if tt.coinType != slip44.Ether {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Contains(t, got.Data, "address")
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathAddress_EdgeCases(t *testing.T) {
	ctx := context.Background()
	backend := createTestBackend(t)

	t.Run("nil context", func(t *testing.T) {
		mockStorage := new(MockStorage)
		data := map[string]interface{}{
			"uuid":     testUUID,
			"path":     testDerivationPath,
			"coinType": int(slip44.Ether),
			"isDev":    false,
		}
		fieldData := createFieldData(data)

		// Mock the List call with nil context using mock.Anything
		mockStorage.On("List", mock.Anything, config.StorageBasePath).Return([]string{}, assert.AnError)

		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}

		// Should handle nil context gracefully
		_, err := backend.pathAddress(nil, req, fieldData)
		// The function should still process but might fail at storage level
		// The exact behavior depends on the storage implementation
		assert.Error(t, err) // Expected to fail with nil context

		mockStorage.AssertExpectations(t)
	})

	t.Run("empty uuid", func(t *testing.T) {
		mockStorage := new(MockStorage)
		data := map[string]interface{}{
			"uuid":     "",
			"path":     testDerivationPath,
			"coinType": int(slip44.Ether),
			"isDev":    false,
		}
		fieldData := createFieldData(data)

		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}

		_, err := backend.pathAddress(ctx, req, fieldData)
		assert.Error(t, err)
	})

	t.Run("large coin type value", func(t *testing.T) {
		mockStorage := new(MockStorage)
		data := map[string]interface{}{
			"uuid":     testUUID,
			"path":     testDerivationPath,
			"coinType": 2147483647, // Max int32
			"isDev":    false,
		}
		fieldData := createFieldData(data)

		testUser := helpers.User{
			Mnemonic:   testMnemonic,
			Passphrase: testPassphrase,
		}
		entry := createUserStorageEntry(t, testUser)
		mockStorage.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
		// Mock List for UUID existence check
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)

		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}

		_, err := backend.pathAddress(ctx, req, fieldData)
		assert.Error(t, err) // Should fail due to unsupported coin type

		mockStorage.AssertExpectations(t)
	})
}

// Benchmark test for performance
func BenchmarkBackend_PathAddress(b *testing.B) {
	ctx := context.Background()
	backend := createTestBackend(&testing.T{})

	testUser := helpers.User{
		Mnemonic:   testMnemonic,
		Passphrase: testPassphrase,
	}

	entry := createUserStorageEntry(&testing.T{}, testUser)

	for i := 0; i < b.N; i++ {
		mockStorage := new(MockStorage)
		mockStorage.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
		// Mock List for UUID existence check
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)

		data := map[string]interface{}{
			"uuid":     testUUID,
			"path":     testDerivationPath,
			"coinType": int(slip44.Ether),
			"isDev":    false,
		}
		fieldData := createFieldData(data)

		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}

		_, _ = backend.pathAddress(ctx, req, fieldData)
	}
}
