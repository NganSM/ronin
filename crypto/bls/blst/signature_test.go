//go:build ((linux && amd64) || (linux && arm64) || (darwin && amd64) || (darwin && arm64) || (windows && amd64)) && !blst_disabled

package blst

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/bls/common"
)

func TestSignVerify(t *testing.T) {
	priv, err := RandKey()
	require.NoError(t, err)
	pub := priv.PublicKey()
	msg := []byte("hello")
	sig := priv.Sign(msg)
	assert.Equal(t, true, sig.Verify(pub, msg), "Signature did not verify")
}

func TestAggregateVerify(t *testing.T) {
	pubkeys := make([]common.PublicKey, 0, 100)
	sigs := make([]common.Signature, 0, 100)
	var msgs [][32]byte
	for i := 0; i < 100; i++ {
		msg := [32]byte{'h', 'e', 'l', 'l', 'o', byte(i)}
		priv, err := RandKey()
		require.NoError(t, err)
		pub := priv.PublicKey()
		sig := priv.Sign(msg[:])
		pubkeys = append(pubkeys, pub)
		sigs = append(sigs, sig)
		msgs = append(msgs, msg)
	}
	aggSig := AggregateSignatures(sigs)
	// skipcq: GO-W1009
	assert.Equal(t, true, aggSig.AggregateVerify(pubkeys, msgs), "Signature did not verify")
}

func TestAggregateVerify_CompressedSignatures(t *testing.T) {
	pubkeys := make([]common.PublicKey, 0, 100)
	sigs := make([]common.Signature, 0, 100)
	var sigBytes [][]byte
	var msgs [][32]byte
	for i := 0; i < 100; i++ {
		msg := [32]byte{'h', 'e', 'l', 'l', 'o', byte(i)}
		priv, err := RandKey()
		require.NoError(t, err)
		pub := priv.PublicKey()
		sig := priv.Sign(msg[:])
		pubkeys = append(pubkeys, pub)
		sigs = append(sigs, sig)
		sigBytes = append(sigBytes, sig.Marshal())
		msgs = append(msgs, msg)
	}
	aggSig := AggregateSignatures(sigs)
	// skipcq: GO-W1009
	assert.Equal(t, true, aggSig.AggregateVerify(pubkeys, msgs), "Signature did not verify")

	aggSig2, err := AggregateCompressedSignatures(sigBytes)
	assert.NoError(t, err)
	assert.Equal(t, aggSig.Marshal(), aggSig2.Marshal(), "Signature did not match up")
}

func TestFastAggregateVerify(t *testing.T) {
	pubkeys := make([]common.PublicKey, 0, 100)
	sigs := make([]common.Signature, 0, 100)
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < 100; i++ {
		priv, err := RandKey()
		require.NoError(t, err)
		pub := priv.PublicKey()
		sig := priv.Sign(msg[:])
		pubkeys = append(pubkeys, pub)
		sigs = append(sigs, sig)
	}
	aggSig := AggregateSignatures(sigs)
	assert.Equal(t, true, aggSig.FastAggregateVerify(pubkeys, msg), "Signature did not verify")

}

func TestVerifyCompressed(t *testing.T) {
	priv, err := RandKey()
	require.NoError(t, err)
	pub := priv.PublicKey()
	msg := []byte("hello")
	sig := priv.Sign(msg)
	assert.Equal(t, true, sig.Verify(pub, msg), "Non compressed signature did not verify")
	assert.Equal(t, true, VerifyCompressed(sig.Marshal(), pub.Marshal(), msg), "Compressed signatures and pubkeys did not verify")
}

func TestVerifySingleSignature_InvalidSignature(t *testing.T) {
	priv, err := RandKey()
	require.NoError(t, err)
	pub := priv.PublicKey()
	msgA := [32]byte{'h', 'e', 'l', 'l', 'o'}
	msgB := [32]byte{'o', 'l', 'l', 'e', 'h'}
	sigA := priv.Sign(msgA[:]).Marshal()
	valid, err := VerifySignature(sigA, msgB, pub)
	assert.NoError(t, err)
	assert.Equal(t, false, valid, "Signature did verify")
}

func TestVerifySingleSignature_ValidSignature(t *testing.T) {
	priv, err := RandKey()
	require.NoError(t, err)
	pub := priv.PublicKey()
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}
	sig := priv.Sign(msg[:]).Marshal()
	valid, err := VerifySignature(sig, msg, pub)
	assert.NoError(t, err)
	assert.Equal(t, true, valid, "Signature did not verify")
}

func TestMultipleSignatureVerification(t *testing.T) {
	pubkeys := make([]common.PublicKey, 0, 100)
	sigs := make([][]byte, 0, 100)
	var msgs [][32]byte
	for i := 0; i < 100; i++ {
		msg := [32]byte{'h', 'e', 'l', 'l', 'o', byte(i)}
		priv, err := RandKey()
		require.NoError(t, err)
		pub := priv.PublicKey()
		sig := priv.Sign(msg[:]).Marshal()
		pubkeys = append(pubkeys, pub)
		sigs = append(sigs, sig)
		msgs = append(msgs, msg)
	}
	verify, err := VerifyMultipleSignatures(sigs, msgs, pubkeys)
	assert.NoError(t, err, "Signature did not verify")
	assert.Equal(t, true, verify, "Signature did not verify")
}

func TestFastAggregateVerify_ReturnsFalseOnEmptyPubKeyList(t *testing.T) {
	var pubkeys []common.PublicKey
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}

	aggSig := NewAggregateSignature()
	assert.Equal(t, false, aggSig.FastAggregateVerify(pubkeys, msg), "Expected FastAggregateVerify to return false with empty input ")
}

func TestEth2FastAggregateVerify(t *testing.T) {
	pubkeys := make([]common.PublicKey, 0, 100)
	sigs := make([]common.Signature, 0, 100)
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}
	for i := 0; i < 100; i++ {
		priv, err := RandKey()
		require.NoError(t, err)
		pub := priv.PublicKey()
		sig := priv.Sign(msg[:])
		pubkeys = append(pubkeys, pub)
		sigs = append(sigs, sig)
	}
	aggSig := AggregateSignatures(sigs)
	assert.Equal(t, true, aggSig.Eth2FastAggregateVerify(pubkeys, msg), "Signature did not verify")

}

func TestEth2FastAggregateVerify_ReturnsFalseOnEmptyPubKeyList(t *testing.T) {
	var pubkeys []common.PublicKey
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}

	aggSig := NewAggregateSignature()
	assert.Equal(t, false, aggSig.Eth2FastAggregateVerify(pubkeys, msg), "Expected Eth2FastAggregateVerify to return false with empty input ")
}

func TestEth2FastAggregateVerify_ReturnsTrueOnG2PointAtInfinity(t *testing.T) {
	var pubkeys []common.PublicKey
	msg := [32]byte{'h', 'e', 'l', 'l', 'o'}

	g2PointAtInfinity := append([]byte{0xC0}, make([]byte, 95)...)
	aggSig, err := SignatureFromBytes(g2PointAtInfinity)
	require.NoError(t, err)
	assert.Equal(t, true, aggSig.Eth2FastAggregateVerify(pubkeys, msg))
}

func TestSignatureFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		err   error
	}{
		{
			name: "Nil",
			err:  errors.New("signature must be 96 bytes"),
		},
		{
			name:  "Empty",
			input: []byte{},
			err:   errors.New("signature must be 96 bytes"),
		},
		{
			name:  "Short",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   errors.New("signature must be 96 bytes"),
		},
		{
			name:  "Long",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   errors.New("signature must be 96 bytes"),
		},
		{
			name:  "Bad",
			input: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			err:   errors.New("could not unmarshal bytes into signature"),
		},
		{
			name:  "Good",
			input: []byte{0xab, 0xb0, 0x12, 0x4c, 0x75, 0x74, 0xf2, 0x81, 0xa2, 0x93, 0xf4, 0x18, 0x5c, 0xad, 0x3c, 0xb2, 0x26, 0x81, 0xd5, 0x20, 0x91, 0x7c, 0xe4, 0x66, 0x65, 0x24, 0x3e, 0xac, 0xb0, 0x51, 0x00, 0x0d, 0x8b, 0xac, 0xf7, 0x5e, 0x14, 0x51, 0x87, 0x0c, 0xa6, 0xb3, 0xb9, 0xe6, 0xc9, 0xd4, 0x1a, 0x7b, 0x02, 0xea, 0xd2, 0x68, 0x5a, 0x84, 0x18, 0x8a, 0x4f, 0xaf, 0xd3, 0x82, 0x5d, 0xaf, 0x6a, 0x98, 0x96, 0x25, 0xd7, 0x19, 0xcc, 0xd2, 0xd8, 0x3a, 0x40, 0x10, 0x1f, 0x4a, 0x45, 0x3f, 0xca, 0x62, 0x87, 0x8c, 0x89, 0x0e, 0xca, 0x62, 0x23, 0x63, 0xf9, 0xdd, 0xb8, 0xf3, 0x67, 0xa9, 0x1e, 0x84},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := SignatureFromBytes(test.input)
			if test.err != nil {
				assert.NotEqual(t, nil, err, "No error returned")
				assert.ErrorContains(t, test.err, err.Error(), "Unexpected error returned")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, 0, bytes.Compare(res.Marshal(), test.input))
			}
		})
	}
}

func TestMultipleSignatureFromBytes(t *testing.T) {
	tests := []struct {
		name  string
		input [][]byte
		err   error
	}{
		{
			name: "Nil",
			err:  errors.New("0 signatures provided to the method"),
		},
		{
			name:  "Empty",
			input: [][]byte{},
			err:   errors.New("0 signatures provided to the method"),
		},
		{
			name:  "Short",
			input: [][]byte{{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			err:   errors.New("signature must be 96 bytes"),
		},
		{
			name:  "Long",
			input: [][]byte{{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			err:   errors.New("signature must be 96 bytes"),
		},
		{
			name: "Bad",
			input: [][]byte{{0x8f, 0xc0, 0xb4, 0x9e, 0x2e, 0xac, 0x50, 0x86, 0xe2, 0xe2, 0xaa, 0xf, 0xdc, 0x54, 0x23, 0x51, 0x6, 0xd8, 0x29, 0xf5, 0xae, 0x3, 0x5d, 0xb8, 0x31, 0x4d, 0x26, 0x3, 0x48, 0x18, 0xb9, 0x1f, 0x6b, 0xd7, 0x86, 0xb4, 0xa2, 0x69, 0xc7, 0xe7, 0xf5, 0xc0, 0x93, 0x19, 0x6e, 0xfd, 0x33, 0xb8, 0x1, 0xe1, 0x1f, 0x4e, 0xb4, 0xb1, 0xa0, 0x1, 0x30, 0x48, 0x8a, 0x6c, 0x97, 0x29, 0xd6, 0xcb, 0x1c, 0x45, 0xef, 0x87, 0xba, 0x4f, 0xce, 0x22, 0x84, 0x48, 0xad, 0x16, 0xf7, 0x5c, 0xb2, 0xa8, 0x34, 0xb9, 0xee, 0xb8, 0xbf, 0xe5, 0x58, 0x2c, 0x44, 0x7b, 0x1f, 0x9c, 0x22, 0x26, 0x3a, 0x22},
				{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			err: errors.New("could not unmarshal bytes into signature"),
		},
		{
			name: "Good",
			input: [][]byte{
				{0xab, 0xb0, 0x12, 0x4c, 0x75, 0x74, 0xf2, 0x81, 0xa2, 0x93, 0xf4, 0x18, 0x5c, 0xad, 0x3c, 0xb2, 0x26, 0x81, 0xd5, 0x20, 0x91, 0x7c, 0xe4, 0x66, 0x65, 0x24, 0x3e, 0xac, 0xb0, 0x51, 0x00, 0x0d, 0x8b, 0xac, 0xf7, 0x5e, 0x14, 0x51, 0x87, 0x0c, 0xa6, 0xb3, 0xb9, 0xe6, 0xc9, 0xd4, 0x1a, 0x7b, 0x02, 0xea, 0xd2, 0x68, 0x5a, 0x84, 0x18, 0x8a, 0x4f, 0xaf, 0xd3, 0x82, 0x5d, 0xaf, 0x6a, 0x98, 0x96, 0x25, 0xd7, 0x19, 0xcc, 0xd2, 0xd8, 0x3a, 0x40, 0x10, 0x1f, 0x4a, 0x45, 0x3f, 0xca, 0x62, 0x87, 0x8c, 0x89, 0x0e, 0xca, 0x62, 0x23, 0x63, 0xf9, 0xdd, 0xb8, 0xf3, 0x67, 0xa9, 0x1e, 0x84},
				{0xb7, 0x86, 0xe5, 0x7, 0x43, 0xe2, 0x53, 0x6c, 0x15, 0x51, 0x9c, 0x6, 0x2a, 0xa7, 0xe5, 0x12, 0xf9, 0xb7, 0x77, 0x93, 0x3f, 0x55, 0xb3, 0xaf, 0x38, 0xf7, 0x39, 0xe4, 0x84, 0x6d, 0x88, 0x44, 0x52, 0x77, 0x65, 0x42, 0x95, 0xd9, 0x79, 0x93, 0x7e, 0xc8, 0x12, 0x60, 0xe3, 0x24, 0xea, 0x8, 0x10, 0x52, 0xcd, 0xd2, 0x7f, 0x5d, 0x25, 0x3a, 0xa8, 0x9b, 0xb7, 0x65, 0xa9, 0x31, 0xea, 0x7c, 0x85, 0x13, 0x53, 0xc0, 0xa3, 0x88, 0xd1, 0xa5, 0x54, 0x85, 0x2, 0x2d, 0xf8, 0xa1, 0xd7, 0xc1, 0x60, 0x58, 0x93, 0xec, 0x7c, 0xf9, 0x33, 0x43, 0x4, 0x48, 0x40, 0x97, 0xef, 0x67, 0x2a, 0x27},
				{0xb2, 0x12, 0xd0, 0xec, 0x46, 0x76, 0x6b, 0x24, 0x71, 0x91, 0x2e, 0xa8, 0x53, 0x9a, 0x48, 0xa3, 0x78, 0x30, 0xc, 0xe8, 0xf0, 0x86, 0xa3, 0x68, 0xec, 0xe8, 0x96, 0x43, 0x34, 0xda, 0xf, 0xf4, 0x65, 0x48, 0xbb, 0xe0, 0x92, 0xa1, 0x8, 0x12, 0x18, 0x46, 0xe6, 0x4a, 0xd6, 0x92, 0x88, 0xe, 0x2, 0xf5, 0xf3, 0x2a, 0x96, 0xb1, 0x4, 0xf1, 0x11, 0xa9, 0x92, 0x79, 0x52, 0x0, 0x64, 0x34, 0xeb, 0x25, 0xe, 0xf4, 0x29, 0x6b, 0x39, 0x4e, 0x28, 0x78, 0xfe, 0x25, 0xa3, 0xc0, 0x88, 0x5a, 0x40, 0xfd, 0x71, 0x37, 0x63, 0x79, 0xcd, 0x6b, 0x56, 0xda, 0xee, 0x91, 0x26, 0x72, 0xfc, 0xbc},
				{0x8f, 0xc0, 0xb4, 0x9e, 0x2e, 0xac, 0x50, 0x86, 0xe2, 0xe2, 0xaa, 0xf, 0xdc, 0x54, 0x23, 0x51, 0x6, 0xd8, 0x29, 0xf5, 0xae, 0x3, 0x5d, 0xb8, 0x31, 0x4d, 0x26, 0x3, 0x48, 0x18, 0xb9, 0x1f, 0x6b, 0xd7, 0x86, 0xb4, 0xa2, 0x69, 0xc7, 0xe7, 0xf5, 0xc0, 0x93, 0x19, 0x6e, 0xfd, 0x33, 0xb8, 0x1, 0xe1, 0x1f, 0x4e, 0xb4, 0xb1, 0xa0, 0x1, 0x30, 0x48, 0x8a, 0x6c, 0x97, 0x29, 0xd6, 0xcb, 0x1c, 0x45, 0xef, 0x87, 0xba, 0x4f, 0xce, 0x22, 0x84, 0x48, 0xad, 0x16, 0xf7, 0x5c, 0xb2, 0xa8, 0x34, 0xb9, 0xee, 0xb8, 0xbf, 0xe5, 0x58, 0x2c, 0x44, 0x7b, 0x1f, 0x9c, 0x22, 0x26, 0x3a, 0x22},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := MultipleSignaturesFromBytes(test.input)
			if test.err != nil {
				assert.NotEqual(t, nil, err, "No error returned")
				assert.ErrorContains(t, test.err, err.Error(), "Unexpected error returned")
			} else {
				assert.NoError(t, err)
				for i, s := range res {
					assert.Equal(t, 0, bytes.Compare(s.Marshal(), test.input[i]))
				}
			}
		})
	}
}

func TestCopy(t *testing.T) {
	priv, err := RandKey()
	require.NoError(t, err)
	key, ok := priv.(*bls12SecretKey)
	require.Equal(t, true, ok)

	signatureA := &Signature{s: new(blstSignature).Sign(key.p, []byte("foo"), dst)}
	signatureB, ok := signatureA.Copy().(*Signature)
	require.Equal(t, true, ok)

	if signatureA == signatureB {
		t.Fatalf("%#v expected not equal to %#v", signatureA, signatureB)
	}

	if signatureA.s == signatureB.s {
		t.Fatalf("%#v expected not equal to %#v", signatureA.s, signatureB.s)
	}
	assert.Equal(t, signatureA, signatureB)

	signatureA.s.Sign(key.p, []byte("bar"), dst)
	assert.NotEqual(t, signatureA, signatureB)
}
