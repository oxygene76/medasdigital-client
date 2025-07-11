package blockchain

import (
	"fmt"
	
	"github.com/cosmos/cosmos-sdk/codec/address"
)

// AddressCodec interface for address operations
type AddressCodec interface {
	StringToBytes(text string) ([]byte, error)
	BytesToString(bz []byte) (string, error)
}

// Bech32AddressCodec implements AddressCodec using cosmos-sdk address codec
type Bech32AddressCodec struct {
	codec address.Codec
}

// NewBech32AddressCodec creates a new Bech32 address codec
func NewBech32AddressCodec(prefix string) AddressCodec {
	return &Bech32AddressCodec{
		codec: address.NewBech32Codec(prefix),
	}
}

// StringToBytes converts address string to bytes
func (b *Bech32AddressCodec) StringToBytes(text string) ([]byte, error) {
	return b.codec.StringToBytes(text)
}

// BytesToString converts address bytes to string
func (b *Bech32AddressCodec) BytesToString(bz []byte) (string, error) {
	return b.codec.BytesToString(bz)
}

// GetSDKCodec returns the underlying SDK codec for keyring operations
func (b *Bech32AddressCodec) GetSDKCodec() address.Codec {
	return b.codec
}

// ValidateAddress validates a bech32 address
func ValidateAddress(codec AddressCodec, addr string) error {
	_, err := codec.StringToBytes(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	return nil
}
