package blockchain

import (
	"fmt"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// AddressCodec interface for address operations
type AddressCodec interface {
	StringToBytes(text string) ([]byte, error)
	BytesToString(bz []byte) (string, error)
}

// Bech32AddressCodec implements AddressCodec using SDK bech32 functions
type Bech32AddressCodec struct {
	prefix string
}

// NewBech32AddressCodec creates a new Bech32 address codec
func NewBech32AddressCodec(prefix string) AddressCodec {
	return &Bech32AddressCodec{
		prefix: prefix,
	}
}

// StringToBytes converts address string to bytes
func (b *Bech32AddressCodec) StringToBytes(text string) ([]byte, error) {
	_, bz, err := bech32.DecodeAndConvert(text)
	return bz, err
}

// BytesToString converts address bytes to string
func (b *Bech32AddressCodec) BytesToString(bz []byte) (string, error) {
	return bech32.ConvertAndEncode(b.prefix, bz)
}

// GetPrefix returns the bech32 prefix
func (b *Bech32AddressCodec) GetPrefix() string {
	return b.prefix
}

// ValidateAddress validates a bech32 address
func ValidateAddress(codec AddressCodec, addr string) error {
	_, err := codec.StringToBytes(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	return nil
}

// GetSDKCodec returns a simple interface for keyring compatibility
func (b *Bech32AddressCodec) GetSDKCodec() AddressCodec {
	return b
}

// Helper function to create AccAddress from string
func StringToAccAddress(addrStr string) (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(addrStr)
	if err != nil {
		return nil, fmt.Errorf("invalid account address: %w", err)
	}
	return addr, nil
}

// NewBech32Codec creates a new codec - compatibility function
func NewBech32Codec(prefix string) AddressCodec {
	return NewBech32AddressCodec(prefix)
}
