package evm

import "errors"

var (
	ErrInvalidECDSAPublicKey = errors.New("invalid ECDSA public key")
	ErrInvalidPayloadData    = errors.New("invalid payload data")
)
