package tron

import "errors"

// Static error variables to avoid dynamic error creation
var (
	ErrInvalidDerivationPath      = errors.New("invalid derivation path")
	ErrInvalidRawData             = errors.New("invalid raw transaction data")
	ErrUnsupportedTransactionType = errors.New("unsupported transaction type")
)
