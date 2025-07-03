package lib

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcutil/hdkeychain"
	bip32 "github.com/tyler-smith/go-bip32"
)

const (
	// DerivationPathCapacity is the initial capacity for derivation path slices
	DerivationPathCapacity = 8
)

// Static error variables to avoid dynamic error creation
var (
	ErrEmptyDerivationPath = errors.New("empty derivation path")
	ErrAmbiguousPath       = errors.New("ambiguous path: use 'm/' prefix for absolute paths, " +
		"or no leading '/' for relative ones")
	ErrInvalidComponent            = errors.New("invalid component in derivation path")
	ErrComponentOutOfRange         = errors.New("component out of allowed range")
	ErrComponentOutOfHardenedRange = errors.New("component out of allowed hardened range")
)

// getDefaultRootDerivationPath returns the default root derivation path.
// This replaces the global variable with a function to avoid linter issues.
func getDefaultRootDerivationPath() derivationPath {
	return derivationPath{0x80000000 + 44, 0x80000000 + 60, 0x80000000 + 0, 0}
}

// DerivationPath represents the computer friendly version of a hierarchical
// deterministic wallet account derivaion path.
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// The BIP-44 spec https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki
// defines that the `purpose` be 44' (or 0x8000002C) for crypto currencies, and
// SLIP-44 https://github.com/satoshilabs/slips/blob/master/slip-0044.md assigns
// the `coin_type` 60' (or 0x8000003C) to Ethereum.
type derivationPath []uint32

// DerivePrivateKey derives the private key of the derivation path.
func DerivePrivateKey(seed []byte, path string, _ bool) (*btcec.PrivateKey, error) {
	// parse derivation path
	deriavtionPath, err := parseDerivationPath(path)
	if err != nil {
		return nil, err
	}

	key, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	for _, n := range deriavtionPath {
		key, err = key.NewChildKey(n)
		if err != nil {
			return nil, err
		}
	}

	privKey, err := hdkeychain.NewKeyFromString(key.B58Serialize())
	if err != nil {
		return nil, err
	}

	return privKey.ECPrivKey()
}

// ParseDerivationPath converts a user specified derivation path string to the
// internal binary representation.
//
// Full derivation paths need to start with the `m/` prefix, relative derivation
// paths (which will get appended to the default root path) must not have prefixes
// in front of the first element. Whitespace is ignored.
func parseDerivationPath(path string) (derivationPath, error) {
	// Pre-allocate result slice with estimated capacity
	result := make(derivationPath, 0, DerivationPathCapacity)

	// Handle absolute or relative paths
	components := strings.Split(path, "/")
	switch {
	case len(components) == 0:
		return nil, ErrEmptyDerivationPath

	case strings.TrimSpace(components[0]) == "":
		return nil, ErrAmbiguousPath

	case strings.TrimSpace(components[0]) == "m":
		components = components[1:]

	default:
		result = append(result, getDefaultRootDerivationPath()...)
	}
	// All remaining components are relative, append one by one
	if len(components) == 0 {
		return nil, ErrEmptyDerivationPath // Empty relative paths
	}
	for _, component := range components {
		// Ignore any user added whitespace
		component = strings.TrimSpace(component)
		var value uint32

		// Handle hardened paths
		if strings.HasSuffix(component, "'") {
			value = 0x80000000
			component = strings.TrimSpace(strings.TrimSuffix(component, "'"))
		}
		// Handle the non hardened component
		bigval, ok := new(big.Int).SetString(component, 0)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrInvalidComponent, component)
		}
		maxRange := math.MaxUint32 - value
		if bigval.Sign() < 0 || bigval.Cmp(big.NewInt(int64(maxRange))) > 0 {
			if value == 0 {
				return nil, fmt.Errorf("%w [0, %d]: %v", ErrComponentOutOfRange, maxRange, bigval)
			}
			return nil, fmt.Errorf("%w [0, %d]: %v", ErrComponentOutOfHardenedRange, maxRange, bigval)
		}
		value += uint32(bigval.Uint64())

		// Append and repeat
		result = append(result, value)
	}
	return result, nil
}
