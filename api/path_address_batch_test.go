package api

import (
	"context"
	"encoding/json"
	"log/slog"
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

// MockStorageBatch implements logical.Storage for testing
// (pattern from path_address_test.go)
type MockStorageBatch struct {
	mock.Mock
}

func (m *MockStorageBatch) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorageBatch) Get(ctx context.Context, key string) (*logical.StorageEntry, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*logical.StorageEntry), args.Error(1)
}

func (m *MockStorageBatch) Put(ctx context.Context, entry *logical.StorageEntry) error {
	args := m.Called(ctx, entry)
	return args.Error(0)
}

func (m *MockStorageBatch) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func createBatchFieldData(data map[string]interface{}) *framework.FieldData {
	schema := map[string]*framework.FieldSchema{
		"uuid": {
			Type:        framework.TypeString,
			Description: "UUID of user",
		},
		"pathTemplate": {
			Type:        framework.TypeString,
			Description: "Templated derivation path",
		},
		"coinType": {
			Type:        framework.TypeInt,
			Description: "Coin type",
		},
		"isDev": {
			Type:        framework.TypeBool,
			Description: "Development mode flag",
		},
		"startIndex": {
			Type:        framework.TypeInt,
			Description: "Start index",
		},
		"count": {
			Type:        framework.TypeInt,
			Description: "Count",
		},
	}
	return &framework.FieldData{
		Raw:    data,
		Schema: schema,
	}
}

func createBatchTestBackend(_ *testing.T) *Backend {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return &Backend{
		logger: logger,
	}
}

func createUserStorageEntryBatch(t *testing.T, uuid, mnemonic, passphrase string) *logical.StorageEntry {
	user := helpers.User{
		UUID:       uuid,
		Mnemonic:   mnemonic,
		Passphrase: passphrase,
	}
	data, err := json.Marshal(user)
	require.NoError(t, err)
	return &logical.StorageEntry{
		Key:   config.StorageBasePath + uuid,
		Value: data,
	}
}

func TestBackend_PathAddressBatch(t *testing.T) {
	ctx := context.Background()
	const (
		testUUID     = "test-uuid-batch"
		testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
		testPass     = "test-passphrase"
	)

	mockStorage := new(MockStorageBatch)
	entry := createUserStorageEntryBatch(t, testUUID, testMnemonic, testPass)
	mockStorage.On("Get", ctx, config.StorageBasePath+testUUID).Return(entry, nil)
	mockStorage.On("List", ctx, config.StorageBasePath).Return([]string{testUUID}, nil)

	backend := createBatchTestBackend(t)
	fieldData := createBatchFieldData(map[string]interface{}{
		"uuid":         testUUID,
		"pathTemplate": "m/44'/60'/0'/0/%d",
		"coinType":     60, // Ether
		"isDev":        false,
		"startIndex":   0,
		"count":        3,
	})

	req := &logical.Request{
		Storage: mockStorage,
		Data: map[string]interface{}{
			"uuid":         testUUID,
			"pathTemplate": "m/44'/60'/0'/0/%d",
			"coinType":     60,
			"isDev":        false,
			"startIndex":   0,
			"count":        3,
		},
	}

	resp, err := backend.pathAddressBatch(ctx, req, fieldData)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, resp.Data, "addresses")
	addresses, ok := resp.Data["addresses"].(map[string]string)
	assert.True(t, ok)
	assert.Len(t, addresses, 3)
	for _, addr := range addresses {
		assert.NotEmpty(t, addr)
	}
}
