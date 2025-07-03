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

	"github.com/payment-system/dq-vault/api/helpers"
	"github.com/payment-system/dq-vault/config"
	"github.com/payment-system/dq-vault/lib/slip44"
)

// Test constants for sign tests
const (
	signTestUUID = "test-uuid-123"
	signTestDerivationPath = "m/44'/60'/0'/0/0"
	signTestValidMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	signTestPassphrase = "test-passphrase"
	signTestPayload = `{"nonce":42,"value":1000000000000000000,"gasLimit":21000,"gasPrice":20000000000,"to":"0x742d35Cc6634C0532925a3b8D359A5C5119e32C8","data":"0x","chainId":1}`
	signTestInvalidPayload = `{"invalid": "json"}`
	signTestMalformedPayload = `{invalid json}`
)

// MockStorage implements logical.Storage for testing (reusing pattern from previous tests)
type MockStorageSign struct {
	mock.Mock
}

func (m *MockStorageSign) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorageSign) Get(ctx context.Context, key string) (*logical.StorageEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logical.StorageEntry), args.Error(1)
}

func (m *MockStorageSign) Put(ctx context.Context, entry *logical.StorageEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStorageSign) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Helper function to create a proper framework.FieldData for sign endpoint
func createSignFieldData(data map[string]interface{}) *framework.FieldData {
	schema := map[string]*framework.FieldSchema{
		"uuid": {
			Type:        framework.TypeString,
			Description: "User UUID",
		},
		"path": {
			Type:        framework.TypeString,
			Description: "Derivation path",
		},
		"coinType": {
			Type:        framework.TypeInt,
			Description: "Coin type",
		},
		"payload": {
			Type:        framework.TypeString,
			Description: "Transaction payload",
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

// Helper function to create test backend for sign tests
func createSignTestBackend(t *testing.T) *backend {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &backend{
		logger: logger,
	}
}

// Helper function to create user storage entry
func createUserStorageEntrySign(uuid, username, mnemonic, passphrase string) *logical.StorageEntry {
	user := helpers.User{
		Username:   username,
		UUID:       uuid,
		Mnemonic:   mnemonic,
		Passphrase: passphrase,
	}
	
	userData, _ := json.Marshal(user)
	return &logical.StorageEntry{
		Key:   config.StorageBasePath + uuid,
		Value: userData,
	}
}

func TestBackend_PathSign(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name           string
		fieldData      map[string]interface{}
		setupStorage   func(*MockStorageSign)
		want           *logical.Response
		wantErr        bool
		wantStatusCode int
		wantErrMsg     string
	}{
		{
			name: "successful ethereum transaction signing",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"signature": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful signing with development mode",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    true,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"signature": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "bitshares coin type with path override",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Bitshares),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true, // Bitshares doesn't have adapter
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing uuid field",
			fieldData: map[string]interface{}{
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing path field",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing coinType field",
			fieldData: map[string]interface{}{
				"uuid":    signTestUUID,
				"path":    signTestDerivationPath,
				"payload": signTestPayload,
				"isDev":   false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check since coinType defaults to 0
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true, // Bitcoin (coinType 0) doesn't have adapter
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing payload field",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true, // Empty payload will cause adapter error
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "empty uuid",
			fieldData: map[string]interface{}{
				"uuid":     "",
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantErrMsg:     "Provide a valid UUID",
		},
		{
			name: "empty derivation path",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     "",
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// No storage expectations since validation should fail first
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantErrMsg:     "Provide a valid path",
		},
		{
			name: "uuid does not exist",
			fieldData: map[string]interface{}{
				"uuid":     "nonexistent-uuid",
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List to return empty list (UUID doesn't exist)
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
			wantErrMsg:     "UUID does not exists",
		},
		{
			name: "storage get error",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get to return error
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(nil, assert.AnError)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid json in storage",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get to return invalid JSON
				invalidEntry := &logical.StorageEntry{
					Key:   config.StorageBasePath + signTestUUID,
					Value: []byte("{invalid json}"),
				}
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(invalidEntry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "empty mnemonic in user data",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get with empty mnemonic
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", "", signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "unsupported coin type",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": 99999, // Unsupported coin type
				"payload":  signTestPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid transaction payload",
			fieldData: map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  signTestMalformedPayload,
				"isDev":    false,
			},
			setupStorage: func(ms *MockStorageSign) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
				// Mock Get for retrieving user data
				userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
				ms.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStorage := new(MockStorageSign)
			backend := createSignTestBackend(t)
			
			// Setup storage expectations
			if tt.setupStorage != nil {
				tt.setupStorage(mockStorage)
			}
			
			// Setup field data
			fieldData := createSignFieldData(tt.fieldData)
			
			// Create request with mock storage
			req := &logical.Request{
				Storage: mockStorage,
				Data:    tt.fieldData,
			}
			
			// Execute
			got, err := backend.pathSign(ctx, req, fieldData)
			
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
					assert.Contains(t, got.Data, "signature")
					assert.NotEmpty(t, got.Data["signature"])
					// Verify signature is a string
					signature, ok := got.Data["signature"].(string)
					assert.True(t, ok)
					assert.NotEmpty(t, signature)
				}
			}
			
			// Verify all mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathSign_CoinTypeSpecific(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name     string
		coinType int
		wantErr  bool
	}{
		{
			name:     "ethereum_supported",
			coinType: int(slip44.Ether),
			wantErr:  false,
		},
		{
			name:     "bitcoin_not_supported",
			coinType: int(slip44.Bitcoin),
			wantErr:  true,
		},
		{
			name:     "bitshares_not_supported",
			coinType: int(slip44.Bitshares),
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorageSign)
			backend := createSignTestBackend(t)
			
			// Setup storage expectations
			mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
			userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
			mockStorage.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			
			fieldData := createSignFieldData(map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": tt.coinType,
				"payload":  signTestPayload,
				"isDev":    false,
			})
			
			req := &logical.Request{
				Storage: mockStorage,
				Data: map[string]interface{}{
					"uuid":     signTestUUID,
					"path":     signTestDerivationPath,
					"coinType": tt.coinType,
					"payload":  signTestPayload,
					"isDev":    false,
				},
			}
			
			got, err := backend.pathSign(ctx, req, fieldData)
			
			// For unsupported coin types, we expect an error
			// Currently only Ether is supported by the EVM adapter
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Contains(t, got.Data, "signature")
			}
			
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathSign_PayloadValidation(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name        string
		payload     string
		expectError bool
	}{
		{
			name:        "valid_ether_transfer",
			payload:     `{"nonce":42,"value":1000000000000000000,"gasLimit":21000,"gasPrice":20000000000,"to":"0x742d35Cc6634C0532925a3b8D359A5C5119e32C8","data":"0x","chainId":1}`,
			expectError: false,
		},
		{
			name:        "empty_payload",
			payload:     "",
			expectError: true,
		},
		{
			name:        "malformed_json",
			payload:     `{invalid json`,
			expectError: true,
		},
		{
			name:        "invalid_transaction_structure",
			payload:     `{"invalid": "structure"}`,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorageSign)
			backend := createSignTestBackend(t)
			
			// Setup storage expectations
			mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
			userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
			mockStorage.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
			
			fieldData := createSignFieldData(map[string]interface{}{
				"uuid":     signTestUUID,
				"path":     signTestDerivationPath,
				"coinType": int(slip44.Ether),
				"payload":  tt.payload,
				"isDev":    false,
			})
			
			req := &logical.Request{
				Storage: mockStorage,
				Data: map[string]interface{}{
					"uuid":     signTestUUID,
					"path":     signTestDerivationPath,
					"coinType": int(slip44.Ether),
					"payload":  tt.payload,
					"isDev":    false,
				},
			}
			
			got, err := backend.pathSign(ctx, req, fieldData)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Contains(t, got.Data, "signature")
			}
			
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathSign_EdgeCases(t *testing.T) {
	ctx := context.Background()
	backend := createSignTestBackend(t)
	
	t.Run("nil_context", func(t *testing.T) {
		mockStorage := new(MockStorageSign)
		data := map[string]interface{}{
			"uuid":     signTestUUID,
			"path":     signTestDerivationPath,
			"coinType": int(slip44.Ether),
			"payload":  signTestPayload,
			"isDev":    false,
		}
		fieldData := createSignFieldData(data)
		
		// Mock with nil context using mock.Anything
		mockStorage.On("List", mock.Anything, config.StorageBasePath).Return([]string{}, assert.AnError)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		_, err := backend.pathSign(nil, req, fieldData)
		assert.Error(t, err) // Expected to fail with nil context or UUID not found
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("very_long_derivation_path", func(t *testing.T) {
		mockStorage := new(MockStorageSign)
		longPath := "m/44'/60'/0'/0/" + string(make([]byte, 1000))
		for i := range longPath[14:] {
			longPath = longPath[:14+i] + "1" + longPath[14+i+1:]
		}
		
		data := map[string]interface{}{
			"uuid":     signTestUUID,
			"path":     longPath,
			"coinType": int(slip44.Ether),
			"payload":  signTestPayload,
			"isDev":    false,
		}
		fieldData := createSignFieldData(data)
		
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
		userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
		mockStorage.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		_, err := backend.pathSign(ctx, req, fieldData)
		assert.Error(t, err) // Expected to fail with invalid path
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("large_payload", func(t *testing.T) {
		mockStorage := new(MockStorageSign)
		largePayload := `{"nonce":42,"value":1000000000000000000,"gasLimit":21000,"gasPrice":20000000000,"to":"0x742d35Cc6634C0532925a3b8D359A5C5119e32C8","data":"0x` + string(make([]byte, 10000)) + `","chainId":1}`
		
		data := map[string]interface{}{
			"uuid":     signTestUUID,
			"path":     signTestDerivationPath,
			"coinType": int(slip44.Ether),
			"payload":  largePayload,
			"isDev":    false,
		}
		fieldData := createSignFieldData(data)
		
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
		userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
		mockStorage.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		// This may succeed or fail depending on payload validation
		// The test is mainly to ensure no panic occurs
		assert.NotPanics(t, func() {
			backend.pathSign(ctx, req, fieldData)
		})
		
		mockStorage.AssertExpectations(t)
	})
}

// Benchmark test for performance
func BenchmarkBackend_PathSign(b *testing.B) {
	ctx := context.Background()
	backend := createSignTestBackend(&testing.T{})
	
	for i := 0; i < b.N; i++ {
		mockStorage := new(MockStorageSign)
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{signTestUUID}, nil)
		userEntry := createUserStorageEntrySign(signTestUUID, "test-user", signTestValidMnemonic, signTestPassphrase)
		mockStorage.On("Get", ctx, config.StorageBasePath+signTestUUID).Return(userEntry, nil)
		
		data := map[string]interface{}{
			"uuid":     signTestUUID,
			"path":     signTestDerivationPath,
			"coinType": int(slip44.Ether),
			"payload":  signTestPayload,
			"isDev":    false,
		}
		fieldData := createSignFieldData(data)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		_, _ = backend.pathSign(ctx, req, fieldData)
	}
}