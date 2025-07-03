package api

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// pathInfo corresponds to READ gen/info.
func (b *Backend) pathInfo(_ context.Context, _ *logical.Request,
	_ *framework.FieldData) (*logical.Response, error) {
	return &logical.Response{
		Data: map[string]interface{}{
			"Info": backendHelp,
		},
	}, nil
}
