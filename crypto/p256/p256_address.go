package p256

import (
	"github.com/amolabs/tendermint-amo/crypto"
	"github.com/amolabs/tendermint-amo/encoding/base58"
)

const (
	addressPrefixLength = 2
	addressSize         = addressPrefixLength + 33
)

var (
	addressTestPrefix = []byte{0x0, 0x7F}
	addressMainPrefix = []byte{0x0, 0x6E}
)

func GenAddress(pubKey crypto.PubKey, prefix []byte) crypto.Address {
	r160 := crypto.Ripemd160(crypto.Sha256(pubKey.Bytes()))
	er160 := make([]byte, addressPrefixLength+160/8)
	copy(er160[:addressPrefixLength], prefix)
	copy(er160[addressPrefixLength:], r160)
	checksum := crypto.Sha256(crypto.Sha256(r160))[:4]
	address := append(er160, checksum...)
	encoded := base58.Encode(address)
	return crypto.Address([]byte(encoded))
}

func GenTestNetAddress(pubKey crypto.PubKey) crypto.Address {
	return GenAddress(pubKey, addressTestPrefix)
}

func GenMainNetAddress(pubKey crypto.PubKey) crypto.Address {
	return GenAddress(pubKey, addressMainPrefix)
}
