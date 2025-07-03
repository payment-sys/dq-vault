package rfc6979

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"math/big"
)

const (
	// VLength is the length of v in bytes
	VLength = 32
	// KLength is the length of k in bytes
	KLength = 32
	// NonceLength is the length of nonce in bytes
	NonceLength = 4
	// ZeroValue is the zero value for comparisons
	ZeroValue = 0x00
	// OneValue is the one value for comparisons
	OneValue = 0x01
	// DivisorTwo is the divisor for calculating N/2
	DivisorTwo = 2
	// BitsPerByte is the number of bits per byte
	BitsPerByte = 8
	// BitShift is the bit shift value
	BitShift = 7
)

// getOneInitializer returns the one initializer byte slice
func getOneInitializer() []byte {
	return []byte{OneValue}
}

// HmacSHA256 returns a Hash-based message authentication code
func HmacSHA256(m, k []byte) []byte {
	mac := hmac.New(sha256.New, k)
	mac.Write(m)
	expectedMAC := mac.Sum(nil)
	return expectedMAC
}

// https://tools.ietf.org/html/rfc6979#section-3.2
func generateSecret(priv *ecdsa.PrivateKey, _ func() hash.Hash, hashBytes []byte, test func(*big.Int) bool, nonce int) {
	var hashClone = make([]byte, len(hashBytes))
	copy(hashClone, hashBytes)

	if nonce > 0 {
		nonceA := make([]byte, NonceLength)
		binary.BigEndian.PutUint32(nonceA, uint32(nonce))
		hashClone = append(hashClone, nonceA...)
		hs := sha256.New()
		hs.Write(hashClone)
		hashClone = hs.Sum(nil)
	}

	c := priv.Curve
	x := priv.D.Bytes()
	q := c.Params().N

	// Step B
	v := bytes.Repeat(getOneInitializer(), VLength)

	// Step C (Go zeroes the all allocated memory)
	k := make([]byte, KLength)

	// Step D
	m := append(append(append(v, ZeroValue), x...), hashClone...)
	k = HmacSHA256(m, k)

	// Step E
	v = HmacSHA256(v, k)

	// Step F
	k = HmacSHA256(append(append(append(v, OneValue), x...), hashClone...), k)

	// Step G
	v = HmacSHA256(v, k)

	// Step H1/H2a, ignored as tlen === qlen (256 bit)
	// Step H2b
	v = HmacSHA256(v, k)

	var T = hashToInt(v, c)

	// Step H3, repeat until T is within the interval [1, n - 1]
	for T.Sign() <= 0 || T.Cmp(q) >= 0 || !test(T) {
		k = HmacSHA256(append(v, ZeroValue), k)
		v = HmacSHA256(v, k)
		// Step H1/H2a, again, ignored as tlen === qlen (256 bit)
		// Step H2b again
		v = HmacSHA256(v, k)
		T = hashToInt(v, c)
	}
}

// SignECDSA signs an arbitrary length hash (which should be the result of
// hashing a larger message) using the private key, priv. It returns the
// signature as a pair of integers.
//
// Note that FIPS 186-3 section 4.6 specifies that the hash should be truncated
// to the byte-length of the subgroup. This function does not perform that
// truncation itself.
func SignECDSA(priv *ecdsa.PrivateKey, hashBytes []byte, alg func() hash.Hash, nonce int) (r, s *big.Int, err error) {
	c := priv.Curve
	N := c.Params().N

	var hashClone = make([]byte, len(hashBytes))
	copy(hashClone, hashBytes)

	generateSecret(priv, alg, hashClone, func(k *big.Int) bool {
		inv := new(big.Int).ModInverse(k, N)
		r, _ = c.ScalarBaseMult(k.Bytes())
		r.Mod(r, N)

		if r.Sign() == 0 {
			return false
		}

		e := hashToInt(hashBytes, c)
		s = new(big.Int).Mul(priv.D, r)
		s.Add(s, e)
		s.Mul(s, inv)
		s.Mod(s, N)

		return s.Sign() != 0
	}, nonce)

	// Enforce low S values, see bip62: 'low s values in signatures'
	if s.Cmp(new(big.Int).Div(N, big.NewInt(DivisorTwo))) == 1 {
		s.Sub(N, s)
	}

	return r, s, nil
}

// copied from crypto/ecdsa
func hashToInt(hashBytes []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + BitShift) / BitsPerByte
	if len(hashBytes) > orderBytes {
		hashBytes = hashBytes[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hashBytes)
	excess := len(hashBytes)*BitsPerByte - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}
