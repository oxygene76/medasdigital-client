package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	
	itypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// Codec handles encoding/decoding for blockchain operations
type Codec struct {
	marshaler         codec.Codec
	interfaceRegistry types.InterfaceRegistry
	addressCodec      AddressCodec
	config            CodecConfig
}

// CodecConfig configuration for codec
type CodecConfig struct {
	Bech32Prefix   string
	UseProtobuf    bool
	UseLegacyAmino bool
}

// NewCodec creates a new codec with default configuration
func NewCodec() *Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	RegisterInterfaces(interfaceRegistry)
	
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	
	return &Codec{
		marshaler:         marshaler,
		interfaceRegistry: interfaceRegistry,
		addressCodec:      NewBech32AddressCodec("medas"),
		config: CodecConfig{
			Bech32Prefix:   "medas",
			UseProtobuf:    true,
			UseLegacyAmino: true,
		},
	}
}

// NewCodecWithConfig creates a new codec with custom configuration
func NewCodecWithConfig(config CodecConfig) *Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	RegisterInterfaces(interfaceRegistry)
	
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	
	return &Codec{
		marshaler:         marshaler,
		interfaceRegistry: interfaceRegistry,
		addressCodec:      NewBech32AddressCodec(config.Bech32Prefix),
		config:            config,
	}
}

// GetMarshaler returns the marshaler
func (c *Codec) GetMarshaler() codec.Codec {
	return c.marshaler
}

// GetInterfaceRegistry returns the interface registry
func (c *Codec) GetInterfaceRegistry() types.InterfaceRegistry {
	return c.interfaceRegistry
}

// GetAddressCodec returns the address codec
func (c *Codec) GetAddressCodec() AddressCodec {
	return c.addressCodec
}

// SetAddressCodec sets the address codec
func (c *Codec) SetAddressCodec(codec AddressCodec) {
	c.addressCodec = codec
}

// MarshalJSON marshals an object to JSON
func (c *Codec) MarshalJSON(obj interface{}) ([]byte, error) {
	if protoMsg, ok := obj.(proto.Message); ok {
		return c.marshaler.MarshalJSON(protoMsg)
	}
	
	// Fallback to standard JSON marshaling
	return json.Marshal(obj)
}

// UnmarshalJSON unmarshals JSON to an object
func (c *Codec) UnmarshalJSON(data []byte, obj interface{}) error {
	if protoMsg, ok := obj.(proto.Message); ok {
		return c.marshaler.UnmarshalJSON(data, protoMsg)
	}
	
	// Fallback to standard JSON unmarshaling
	return json.Unmarshal(data, obj)
}

// MarshalBinary marshals an object to binary format
func (c *Codec) MarshalBinary(obj interface{}) ([]byte, error) {
	if protoMsg, ok := obj.(proto.Message); ok {
		return c.marshaler.Marshal(protoMsg)
	}
	
	return nil, fmt.Errorf("object does not implement proto.Message")
}

// UnmarshalBinary unmarshals binary data to an object
func (c *Codec) UnmarshalBinary(data []byte, obj interface{}) error {
	if protoMsg, ok := obj.(proto.Message); ok {
		return c.marshaler.Unmarshal(data, protoMsg)
	}
	
	return fmt.Errorf("object does not implement proto.Message")
}

// EncodeAddress encodes address bytes to string
func (c *Codec) EncodeAddress(addr []byte) (string, error) {
	return c.addressCodec.BytesToString(addr)
}

// DecodeAddress decodes address string to bytes
func (c *Codec) DecodeAddress(addr string) ([]byte, error) {
	return c.addressCodec.StringToBytes(addr)
}

// ValidateAddress validates an address string
func (c *Codec) ValidateAddress(addr string) error {
	return ValidateAddress(c.addressCodec, addr)
}

// RegisterInterfaces registers the interfaces for protobuf
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterClient{},
		&MsgStoreAnalysis{},
		&MsgUpdateClient{},
		&MsgDeactivateClient{},
	)
}

// RegisterLegacyAminoCodec registers the legacy amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterClient{}, "clientregistry/MsgRegisterClient", nil)
	cdc.RegisterConcrete(&MsgStoreAnalysis{}, "analysis/MsgStoreAnalysis", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "clientregistry/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgDeactivateClient{}, "clientregistry/MsgDeactivateClient", nil)
}

// ValidateMessage validates a message
func (c *Codec) ValidateMessage(msg interface{}) error {
	// Try to cast to sdk.Msg first
	if sdkMsg, ok := msg.(sdk.Msg); ok {
		return sdkMsg.ValidateBasic()
	}
	
	// If not sdk.Msg, check if it's a proto message with ValidateBasic method
	if validator, ok := msg.(interface{ ValidateBasic() error }); ok {
		return validator.ValidateBasic()
	}
	
	return fmt.Errorf("message does not implement validation interface")
}

// GetTypeURL returns the type URL for a message
func (c *Codec) GetTypeURL(msg proto.Message) string {
	// Use the interface registry to resolve type URL
	if resolver, ok := c.interfaceRegistry.(interface{
		Resolve(typeUrl string) (proto.Message, error)
	}); ok {
		_ = resolver // We have a resolver but need reverse lookup
		// For now, construct type URL manually
		return "/" + proto.MessageName(msg)
	}
	
	// Fallback: construct type URL from message name
	return "/" + proto.MessageName(msg)
}

// MustMarshalJSON marshals to JSON or panics
func (c *Codec) MustMarshalJSON(obj interface{}) []byte {
	data, err := c.MarshalJSON(obj)
	if err != nil {
		panic(err)
	}
	return data
}

// MustUnmarshalJSON unmarshals from JSON or panics
func (c *Codec) MustUnmarshalJSON(data []byte, obj interface{}) {
	if err := c.UnmarshalJSON(data, obj); err != nil {
		panic(err)
	}
}

// MustMarshalBinary marshals to binary or panics
func (c *Codec) MustMarshalBinary(obj interface{}) []byte {
	data, err := c.MarshalBinary(obj)
	if err != nil {
		panic(err)
	}
	return data
}

// MustUnmarshalBinary unmarshals from binary or panics
func (c *Codec) MustUnmarshalBinary(data []byte, obj interface{}) {
	if err := c.UnmarshalBinary(data, obj); err != nil {
		panic(err)
	}
}

// Global codec instance
var globalCodec *Codec

// InitGlobalCodec initializes the global codec
func InitGlobalCodec() {
	globalCodec = NewCodec()
}

// GetGlobalCodec returns the global codec instance
func GetGlobalCodec() *Codec {
	if globalCodec == nil {
		InitGlobalCodec()
	}
	return globalCodec
}

// Convenience functions using global codec

// MarshalJSON marshals using global codec
func MarshalJSON(codec *Codec, obj interface{}) ([]byte, error) {
	if codec == nil {
		codec = GetGlobalCodec()
	}
	return codec.MarshalJSON(obj)
}

// UnmarshalJSON unmarshals using global codec
func UnmarshalJSON(codec *Codec, data []byte, obj interface{}) error {
	if codec == nil {
		codec = GetGlobalCodec()
	}
	return codec.UnmarshalJSON(data, obj)
}

// MustMarshalJSON marshals to JSON or panics using global codec
func MustMarshalJSON(codec *Codec, obj interface{}) []byte {
	if codec == nil {
		codec = GetGlobalCodec()
	}
	return codec.MustMarshalJSON(obj)
}

// MustUnmarshalJSON unmarshals from JSON or panics using global codec
func MustUnmarshalJSON(codec *Codec, data []byte, obj interface{}) {
	if codec == nil {
		codec = GetGlobalCodec()
	}
	codec.MustUnmarshalJSON(data, obj)
}

// EncodeAddress encodes address using global codec
func EncodeAddress(addr []byte) (string, error) {
	return GetGlobalCodec().EncodeAddress(addr)
}

// DecodeAddress decodes address using global codec
func DecodeAddress(addr string) ([]byte, error) {
	return GetGlobalCodec().DecodeAddress(addr)
}
