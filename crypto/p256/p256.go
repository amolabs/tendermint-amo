package p256

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"strings"

	amino "github.com/tendermint/go-amino"

	tmc "github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/crypto/tmhash"
	cmn "github.com/amolabs/tendermint-amo/libs/common"
)

var (
	c = elliptic.P256()
	h = tmc.Sha256
)

var cdc = amino.NewCodec()

const (
	PrivKeyAminoName = "amo/PrivKeyP256"
	PubKeyAminoName  = "amo/PubKeyP256"
	SignatureSize    = 64
	PrivKeyP256Size  = 32
	PubKeyP256Size   = 65
)

type PrivKeyP256 [PrivKeyP256Size]byte
type PubKeyP256 [PubKeyP256Size]byte

func init() {
	cdc.RegisterInterface((*tmc.PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKeyP256{},
		PubKeyAminoName, nil)

	cdc.RegisterInterface((*tmc.PrivKey)(nil), nil)
	cdc.RegisterConcrete(PrivKeyP256{},
		PrivKeyAminoName, nil)
}

func GenPrivKeyFromSecret(secret []byte) PrivKeyP256 {
	privKey32 := h(secret)
	priv := PrivKeyP256{}
	copy(priv[:], privKey32)
	return priv
}

func GenPrivKey() PrivKeyP256 {
	return genPrivKey(tmc.CReader())
}

func genPrivKey(rand io.Reader) PrivKeyP256 {
	p256, err := ecdsa.GenerateKey(c, rand)
	if err != nil {
		panic(err)
	}
	var privKey PrivKeyP256
	copy(privKey[:], p256.D.Bytes())
	return privKey
}

func (privKey PrivKeyP256) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(privKey)
}

func (privKey PrivKeyP256) RawBytes() []byte {
	return privKey[:]
}

func (privKey PrivKeyP256) ToECDSA() *ecdsa.PrivateKey {
	X, Y := c.ScalarBaseMult(privKey[:])
	return &ecdsa.PrivateKey{
		D: new(big.Int).SetBytes(privKey[:]),
		PublicKey: ecdsa.PublicKey{
			Curve: c,
			X:     X,
			Y:     Y,
		},
	}
}

func (privKey PrivKeyP256) Sign(msg []byte) ([]byte, error) {
	priv := privKey.ToECDSA()
	r, s, err := ecdsa.Sign(tmc.CReader(), priv, h(msg))
	if err != nil {
		return nil, err
	}
	rb := r.Bytes()
	sb := s.Bytes()

	// ASN. 1 format: 0x30|length of remaining data|0x02|length of r|r|0x02|length of s|s
	var sig bytes.Buffer

	sig.WriteByte(0x30)

	sig.WriteByte(byte(2+2+len(rb)+len(sb)))

	sig.WriteByte(0x02)
	sig.WriteByte(byte(len(rb)))
	sig.Write(rb)

	sig.WriteByte(0x02)
	sig.WriteByte(byte(len(sb)))
	sig.Write(sb)

	return sig.Bytes(), nil
}

func (privKey PrivKeyP256) PubKey() tmc.PubKey {
	priv := privKey.ToECDSA()
	pubKey := PubKeyP256{0x04}
	copy(pubKey[1:], priv.X.Bytes())
	copy(pubKey[33:], priv.Y.Bytes())
	return pubKey
}

func (privKey PrivKeyP256) Equals(other tmc.PrivKey) bool {
	return bytes.Equal(privKey.Bytes(), other.Bytes())
}

func (privKey *PrivKeyP256) SetBytes(buf []byte) {
	copy(privKey[:], buf)
}

func (privKey PrivKeyP256) String() string {
	return strings.ToUpper(hex.EncodeToString(privKey[:]))
}

func (privKey PrivKeyP256) MarshalJSON() ([]byte, error) {
	data := make([]byte, len(privKey)*2+2)
	data[0] = '"'
	data[len(data)-1] = '"'
	copy(data[1:], privKey.String())
	return data, nil
}

func (privKey *PrivKeyP256) UnmarshalJSON(data []byte) error {
	if len(data) != PrivKeyP256Size*2 + 2 {
		return cmn.NewError("Invalid private key format")
	}
	_, err := hex.Decode(privKey[:], data[1:len(data)-1])
	if err != nil {
		panic(err)
	}
	return nil
}

func (pubKey PubKeyP256) Address() tmc.Address {
	return tmc.Address(tmhash.SumTruncated(pubKey[:]))
}

func (pubKey PubKeyP256) Bytes() []byte {
	return cdc.MustMarshalBinaryBare(pubKey)
}

func (pubKey PubKeyP256) RawBytes() []byte {
	return pubKey[:]
}

func (pubKey PubKeyP256) ToECDSA() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: c,
		X:     new(big.Int).SetBytes(pubKey[1:33]),
		Y:     new(big.Int).SetBytes(pubKey[33:]),
	}
}

func (pubKey PubKeyP256) VerifyBytes(msg []byte, sig []byte) bool {
	if sig[0] != 0x30 {
		return false
	}

	remainLength := sig[1]
	if int(remainLength) != len(sig) - 2 {
		return false
	}

	// parse r
	if sig[2] != 0x02 {
		return false
	}
	rLen := int(sig[3])
	ind := 4 + rLen
	r := sig[4:ind]

	// parse s
	if sig[ind] != 0x02 {
		return false
	}
	ind += 1
	sLen := int(sig[ind])
	s := sig[ind+1:]

	if len(r) != rLen || len(s) != sLen {
		return false
	}

	return ecdsa.Verify(pubKey.ToECDSA(), h(msg), new(big.Int).SetBytes(r), new(big.Int).SetBytes(s))
}

func (pubKey PubKeyP256) Equals(other tmc.PubKey) bool {
	return bytes.Equal(pubKey.Bytes(), other.Bytes())
}

func (pubKey PubKeyP256) String() string {
	return strings.ToUpper(hex.EncodeToString(pubKey[:]))
}

func (pubKey PubKeyP256) MarshalJSON() ([]byte, error) {
	data := make([]byte, len(pubKey)*2+2)
	data[0] = '"'
	data[len(data)-1] = '"'
	copy(data[1:], pubKey.String())
	return data, nil
}

func (pubKey *PubKeyP256) UnmarshalJSON(data []byte) error {
	if len(data) != PubKeyP256Size*2 + 2 {
		return cmn.NewError("Invalid public key format")
	}
 	_, err := hex.Decode(pubKey[:], data[1:len(data)-1])
 	if err != nil {
 		panic(err)
	}
 	return nil
}

var _ tmc.PrivKey = PrivKeyP256{}
var _ json.Marshaler = (*PrivKeyP256)(nil)
var _ json.Unmarshaler = (*PrivKeyP256)(nil)
var _ tmc.PubKey = PubKeyP256{}
var _ json.Marshaler = (*PubKeyP256)(nil)
var _ json.Unmarshaler = (*PubKeyP256)(nil)
