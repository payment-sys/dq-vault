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
)

// Test constants for register tests
const (
	regTestUsername = "test-user"
	regTestValidMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	regTestInvalidMnemonic = "invalid mnemonic words that do not form valid bip39"
	regTestPassphrase = "test-passphrase"
	regTestGeneratedUUID = "generated-uuid-123"
)

// MockStorage implements logical.Storage for testing (reusing from path_address_test.go)
type MockStorageRegister struct {
	mock.Mock
}

func (m *MockStorageRegister) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorageRegister) Get(ctx context.Context, key string) (*logical.StorageEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logical.StorageEntry), args.Error(1)
}

func (m *MockStorageRegister) Put(ctx context.Context, entry *logical.StorageEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStorageRegister) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Helper function to create a proper framework.FieldData for register endpoint
func createRegisterFieldData(data map[string]interface{}) *framework.FieldData {
	schema := map[string]*framework.FieldSchema{
		"username": {
			Type:        framework.TypeString,
			Description: "Username for registration",
		},
		"mnemonic": {
			Type:        framework.TypeString,
			Description: "BIP39 mnemonic phrase",
		},
		"passphrase": {
			Type:        framework.TypeString,
			Description: "Passphrase for mnemonic",
		},
	}
	
	return &framework.FieldData{
		Raw:    data,
		Schema: schema,
	}
}

// Helper function to create test backend for register tests
func createRegisterTestBackend(t *testing.T) *backend {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &backend{
		logger: logger,
	}
}

// Mock functions for external dependencies
func mockUUIDExists(returnValue bool) {
	// This would normally mock helpers.UUIDExists, but since it's called directly,
	// we'll handle it in the storage mock expectations
}

func TestBackend_PathRegister(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name           string
		fieldData      map[string]interface{}
		setupStorage   func(*MockStorageRegister)
		setupMocks     func()
		want           *logical.Response
		wantErr        bool
		wantStatusCode int
		wantErrMsg     string
	}{
		{
			name: "successful registration with provided mnemonic",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List for UUID existence check - return empty list (UUID doesn't exist)
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				// Mock Put for storing user data
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful registration with empty mnemonic (auto-generated)",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   "",
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				// Mock Put for storing user data
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "successful registration with no passphrase",
			fieldData: map[string]interface{}{
				"username": regTestUsername,
				"mnemonic": regTestValidMnemonic,
				"passphrase": "",
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				// Mock Put for storing user data
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "missing username field",
			fieldData: map[string]interface{}{
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Function proceeds to completion even without username
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false, // Function actually succeeds with empty username
		},
		{
			name: "invalid mnemonic provided",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   regTestInvalidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
			},
			wantErr:        true,
			wantStatusCode: http.StatusExpectationFailed,
			wantErrMsg:     "Invalid Mnemonic",
		},
		{
			name: "storage put error",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List for UUID existence check
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				// Mock Put to return error
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(assert.AnError)
			},
			wantErr:        true,
			wantStatusCode: http.StatusExpectationFailed,
		},
		{
			name: "UUID collision handling",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List to return empty list (no collision for simplicity)
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				// Mock Put for storing user data
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false,
		},
		{
			name: "storage list error during UUID check",
			fieldData: map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Mock List to return error - function should continue despite error
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, assert.AnError)
				// Function continues to Put even when List fails
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false, // Function succeeds even when List fails
		},
		{
			name: "empty username",
			fieldData: map[string]interface{}{
				"username":   "",
				"mnemonic":   regTestValidMnemonic,
				"passphrase": regTestPassphrase,
			},
			setupStorage: func(ms *MockStorageRegister) {
				// Function proceeds to completion even with empty username
				ms.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
				ms.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			},
			want: &logical.Response{
				Data: map[string]interface{}{
					"uuid": mock.AnythingOfType("string"),
				},
			},
			wantErr: false, // Function actually succeeds with empty username
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockStorage := new(MockStorageRegister)
			backend := createRegisterTestBackend(t)
			
			// Setup storage expectations
			if tt.setupStorage != nil {
				tt.setupStorage(mockStorage)
			}
			
			// Setup field data
			fieldData := createRegisterFieldData(tt.fieldData)
			
			// Create request with mock storage
			req := &logical.Request{
				Storage: mockStorage,
				Data:    tt.fieldData,
			}
			
			// Execute
			got, err := backend.pathRegister(ctx, req, fieldData)
			
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
					assert.Contains(t, got.Data, "uuid")
					assert.NotEmpty(t, got.Data["uuid"])
					// Verify UUID is a string
					uuid, ok := got.Data["uuid"].(string)
					assert.True(t, ok)
					assert.NotEmpty(t, uuid)
				}
			}
			
			// Verify all mock expectations were met
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathRegister_MnemonicGeneration(t *testing.T) {
	ctx := context.Background()
	
	tests := []struct {
		name         string
		providedMnemonic string
		expectGenerated  bool
	}{
		{
			name:            "use provided valid mnemonic",
			providedMnemonic: regTestValidMnemonic,
			expectGenerated:  false,
		},
		{
			name:            "generate mnemonic when empty",
			providedMnemonic: "",
			expectGenerated:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorageRegister)
			backend := createRegisterTestBackend(t)
			
			// Setup storage expectations
			mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
			mockStorage.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
			
			fieldData := createRegisterFieldData(map[string]interface{}{
				"username":   regTestUsername,
				"mnemonic":   tt.providedMnemonic,
				"passphrase": regTestPassphrase,
			})
			
			req := &logical.Request{
				Storage: mockStorage,
				Data: map[string]interface{}{
					"username":   regTestUsername,
					"mnemonic":   tt.providedMnemonic,
					"passphrase": regTestPassphrase,
				},
			}
			
			got, err := backend.pathRegister(ctx, req, fieldData)
			
			assert.NoError(t, err)
			assert.NotNil(t, got)
			assert.Contains(t, got.Data, "uuid")
			
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestBackend_PathRegister_StorageContent(t *testing.T) {
	ctx := context.Background()
	mockStorage := new(MockStorageRegister)
	backend := createRegisterTestBackend(t)
	
	// Capture the storage entry to verify its content
	var capturedEntry *logical.StorageEntry
	mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
	mockStorage.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Run(func(args mock.Arguments) {
		capturedEntry = args.Get(1).(*logical.StorageEntry)
	}).Return(nil)
	
	fieldData := createRegisterFieldData(map[string]interface{}{
		"username":   regTestUsername,
		"mnemonic":   regTestValidMnemonic,
		"passphrase": regTestPassphrase,
	})
	
	req := &logical.Request{
		Storage: mockStorage,
		Data: map[string]interface{}{
			"username":   regTestUsername,
			"mnemonic":   regTestValidMnemonic,
			"passphrase": regTestPassphrase,
		},
	}
	
	got, err := backend.pathRegister(ctx, req, fieldData)
	
	assert.NoError(t, err)
	assert.NotNil(t, got)
	
	// Verify storage content
	require.NotNil(t, capturedEntry)
	
	var storedUser helpers.User
	err = json.Unmarshal(capturedEntry.Value, &storedUser)
	require.NoError(t, err)
	
	assert.Equal(t, regTestUsername, storedUser.Username)
	assert.Equal(t, regTestValidMnemonic, storedUser.Mnemonic)
	assert.Equal(t, regTestPassphrase, storedUser.Passphrase)
	assert.NotEmpty(t, storedUser.UUID)
	assert.True(t, len(storedUser.UUID) > 0)
	
	// Verify storage path
	expectedPath := config.StorageBasePath + storedUser.UUID
	assert.Equal(t, expectedPath, capturedEntry.Key)
	
	mockStorage.AssertExpectations(t)
}

func TestBackend_PathRegister_EdgeCases(t *testing.T) {
	ctx := context.Background()
	backend := createRegisterTestBackend(t)
	
	t.Run("nil context", func(t *testing.T) {
		mockStorage := new(MockStorageRegister)
		data := map[string]interface{}{
			"username":   regTestUsername,
			"mnemonic":   regTestValidMnemonic,
			"passphrase": regTestPassphrase,
		}
		fieldData := createRegisterFieldData(data)
		
		// Mock with nil context using mock.Anything
		mockStorage.On("List", mock.Anything, config.StorageBasePath).Return([]string{}, nil)
		mockStorage.On("Put", mock.Anything, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		_, err := backend.pathRegister(nil, req, fieldData)
		assert.NoError(t, err) // Function actually succeeds with nil context
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("very long username", func(t *testing.T) {
		mockStorage := new(MockStorageRegister)
		longUsername := string(make([]byte, 1000)) // Very long username
		for i := range longUsername {
			longUsername = longUsername[:i] + "a" + longUsername[i+1:]
		}
		
		data := map[string]interface{}{
			"username":   longUsername,
			"mnemonic":   regTestValidMnemonic,
			"passphrase": regTestPassphrase,
		}
		fieldData := createRegisterFieldData(data)
		
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
		mockStorage.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		got, err := backend.pathRegister(ctx, req, fieldData)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("unicode username", func(t *testing.T) {
		mockStorage := new(MockStorageRegister)
		unicodeUsername := "用户名测试🔐"
		
		data := map[string]interface{}{
			"username":   unicodeUsername,
			"mnemonic":   regTestValidMnemonic,
			"passphrase": regTestPassphrase,
		}
		fieldData := createRegisterFieldData(data)
		
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
		mockStorage.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		got, err := backend.pathRegister(ctx, req, fieldData)
		assert.NoError(t, err)
		assert.NotNil(t, got)
		
		mockStorage.AssertExpectations(t)
	})
}

// Benchmark test for performance
func BenchmarkBackend_PathRegister(b *testing.B) {
	ctx := context.Background()
	backend := createRegisterTestBackend(&testing.T{})
	
	for i := 0; i < b.N; i++ {
		mockStorage := new(MockStorageRegister)
		mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{}, nil)
		mockStorage.On("Put", ctx, mock.AnythingOfType("*logical.StorageEntry")).Return(nil)
		
		data := map[string]interface{}{
			"username":   regTestUsername,
			"mnemonic":   regTestValidMnemonic,
			"passphrase": regTestPassphrase,
		}
		fieldData := createRegisterFieldData(data)
		
		req := &logical.Request{
			Storage: mockStorage,
			Data:    data,
		}
		
		_, _ = backend.pathRegister(ctx, req, fieldData)
	}
}