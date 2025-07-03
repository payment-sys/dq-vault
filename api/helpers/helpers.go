package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/payment-system/dq-vault/config"
	"github.com/rs/xid"
)

// Static error variables to avoid dynamic error creation
var (
	ErrInvalidUUID      = errors.New("provide a valid UUID")
	ErrInvalidPath      = errors.New("provide a valid path")
	ErrUUIDDoesNotExist = errors.New("UUID does not exists")
	ErrUnknownFields    = errors.New("unknown fields provided")
)

// User -- stores data related to user
type User struct {
	Username   string `json:"username"`
	UUID       string `json:"uuid"`
	Mnemonic   string `json:"mnemonic"`
	Passphrase string `json:"passphrase"`
}

// NewUUID returns a globally unique random generated guid
func NewUUID() string {
	return xid.New().String()
}

// ErrMissingField returns a logical response error that prints a consistent
// error message for when a required field is missing.
func ErrMissingField(field string) *logical.Response {
	return logical.ErrorResponse(fmt.Sprintf("missing required field '%s'", field))
}

// ValidationErr returns an error that corresponds to a validation error.
func ValidationErr(msg string) error {
	return logical.CodedError(http.StatusUnprocessableEntity, msg)
}

// ValidateFields verifies that no bad arguments were given to the request.
func ValidateFields(req *logical.Request, data *framework.FieldData) error {
	var unknownFields []string
	for k := range req.Data {
		if _, ok := data.Schema[k]; !ok {
			unknownFields = append(unknownFields, k)
		}
	}

	// Fix SA4010: Use the unknownFields slice properly
	if len(unknownFields) > 0 {
		return fmt.Errorf("%w: %v", ErrUnknownFields, unknownFields)
	}

	return nil
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// New returns an error that formats as the given text.
func New(text string) error {
	return &errorString{text}
}

// ValidateData - validates data provided to create signature
func ValidateData(ctx context.Context, req *logical.Request, uuid, derivationPath string) error {
	// Check if user provided UUID or not
	if uuid == "" {
		return ErrInvalidUUID
	}

	// base check: if derivation path is valid or not
	if derivationPath == "" {
		return ErrInvalidPath
	}

	if !UUIDExists(ctx, req, uuid) {
		return ErrUUIDDoesNotExist
	}
	return nil
}

// UUIDExists checks if uuid exists or not
func UUIDExists(ctx context.Context, req *logical.Request, uuid string) bool {
	vals, err := req.Storage.List(ctx, config.StorageBasePath)
	if err != nil {
		return false
	}

	for _, val := range vals {
		if val == uuid {
			return true
		}
	}
	return false
}
