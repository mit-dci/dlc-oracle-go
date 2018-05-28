package dlcoracle

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/adiabat/btcd/btcec"
	"github.com/adiabat/btcd/chaincfg/chainhash"
)

var (
	bigZero = new(big.Int).SetInt64(0)
)

// GenerateNumericMessage returns a zero-padded message
// for numeric values, LIT expects numeric oracle values
// to be 256-bit
func GenerateNumericMessage(value uint64) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint64(0))
	binary.Write(&buf, binary.BigEndian, uint64(0))
	binary.Write(&buf, binary.BigEndian, uint64(0))
	binary.Write(&buf, binary.BigEndian, value)
	return buf.Bytes()
}

// PublicKeyFromPrivateKey derives the public key to a private key
func PublicKeyFromPrivateKey(privateKey [32]byte) [33]byte {
	var pubKey [33]byte

	_, pk := btcec.PrivKeyFromBytes(btcec.S256(), privateKey[:])
	copy(pubKey[:], pk.SerializeCompressed())
	return pubKey
}

// GenerateOneTimeSigningKey will return a new random private scalar
// to be used when signing a new message
func GenerateOneTimeSigningKey() ([32]byte, error) {
	var privKey [32]byte
	_, err := rand.Read(privKey[:])
	if err != nil {
		return privKey, err
	}
	return privKey, nil
}

// ComputeSignaturePubKey calculates the signature multipled by the generator
// point, for an arbitrary message based on pubkey R and pubkey A.
// Calculates P = pubR - h(msg, pubR)pubA.
// This is used when building settlement transactions and determining the pubkey
// to the oracle's possible signatures beforehand. Can be calculated with just
// public keys, so by anyone.
func ComputeSignaturePubKey(oraclePubA, oraclePubR [33]byte, message []byte) ([33]byte, error) {
	var returnValue [33]byte

	// Hardcode curve
	curve := btcec.S256()

	A, err := btcec.ParsePubKey(oraclePubA[:], curve)
	if err != nil {
		return returnValue, err
	}

	R, err := btcec.ParsePubKey(oraclePubR[:], curve)
	if err != nil {
		return returnValue, err
	}

	// e = Hash(messageType, oraclePubQ)
	var hashInput []byte
	hashInput = append(message, R.X.Bytes()...)
	e := chainhash.HashB(hashInput)

	bigE := new(big.Int).SetBytes(e)

	if bigE.Cmp(curve.N) >= 0 {
		return returnValue, fmt.Errorf("hash of (msg, pubR) too big")
	}

	// e * B
	A.X, A.Y = curve.ScalarMult(A.X, A.Y, e)

	A.Y.Neg(A.Y)

	A.Y.Mod(A.Y, curve.P)

	P := new(btcec.PublicKey)

	// add to R
	P.X, P.Y = curve.Add(A.X, A.Y, R.X, R.Y)
	copy(returnValue[:], P.SerializeCompressed())
	return returnValue, nil
}

// ComputeSignature Computes the signature for an arbitrary message based on two private scalars:
// The one-time signing key and the oracle's private key.
func ComputeSignature(privKey, oneTimeSigningKey [32]byte, message []byte) ([32]byte, error) {
	var empty, s [32]byte

	// Hardcode curve
	curve := btcec.S256()

	bigPriv := new(big.Int).SetBytes(privKey[:])
	privKey = empty
	bigK := new(big.Int).SetBytes(oneTimeSigningKey[:])

	if bigPriv.Cmp(bigZero) == 0 {
		return empty, fmt.Errorf("priv scalar is zero")
	}
	if bigPriv.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("priv scalar is out of bounds")
	}
	if bigK.Cmp(bigZero) == 0 {
		return empty, fmt.Errorf("k scalar is zero")
	}
	if bigK.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("k scalar is out of bounds")
	}

	// re-derive R = kG
	var Rx *big.Int
	Rx, _ = curve.ScalarBaseMult(oneTimeSigningKey[:])

	// Ry is always even.  Make it even if it's not.
	//	if Ry.Bit(0) == 1 {
	//		bigK.Mod(bigK, curve.N)
	//		bigK.Sub(curve.N, bigK)
	//	}

	// e = Hash(r, m)

	e := chainhash.HashB(append(message[:], Rx.Bytes()...))
	bigE := new(big.Int).SetBytes(e)

	// If the hash is bigger than N, fail.  Note that N is
	// FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141
	// So this happens about once every 2**128 signatures.
	if bigE.Cmp(curve.N) >= 0 {
		return empty, fmt.Errorf("hash of (m, R) too big")
	}
	//	fmt.Printf("e: %x\n", e)
	// s = k + e*a
	bigS := new(big.Int)
	// e*a
	bigS.Mul(bigE, bigPriv)
	// k + (e*a)
	bigS.Sub(bigK, bigS)
	bigS.Mod(bigS, curve.N)

	// check if s is 0, and fail if it is.  Can't see how this would happen;
	// looks like it would happen about once every 2**256 signatures
	if bigS.Cmp(bigZero) == 0 {
		str := fmt.Errorf("sig s %v is zero", bigS)
		return empty, str
	}

	// Zero out private key and k in array and bigint form
	// who knows if this really helps...  can't hurt though.
	bigK.SetInt64(0)
	oneTimeSigningKey = empty
	bigPriv.SetInt64(0)

	byteOffset := (256 - bigS.BitLen()) / 8

	copy(s[byteOffset:], bigS.Bytes())

	return s, nil
}
