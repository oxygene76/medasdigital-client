package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	clienttypes "github.com/oxygene76/medasdigital-client/internal/types"
)

// AddressCodec represents the address codec interface for v0.50
type AddressCodec interface {
	StringToBytes(text string) ([]byte, error)
	BytesToString(bz []byte) (string, error)
}

// Codec represents the application codec for MedasDigital client
type Codec struct {
	marshaler         codec.Codec
	legacyAmino       *codec.LegacyAmino
	interfaceRegistry types.InterfaceRegistry
	addressCodec      AddressCodec
}

// NewCodec creates a new codec instance
func NewCodec() *Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	addressCodec := address.NewBech32Codec("medas")

	// Register interfaces
	RegisterInterfaces(interfaceRegistry)
	RegisterLegacyAminoCodec(legacyAmino)

	return &Codec{
		marshaler:         marshaler,
		legacyAmino:       legacyAmino,
		interfaceRegistry: interfaceRegistry,
		addressCodec:      addressCodec,
	}
}

// RegisterInterfaces registers message interfaces for protobuf
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register message types for the clientregistry module
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgRegisterClient{},
		&MsgStoreAnalysis{},
		&MsgUpdateClient{},
		&MsgDeactivateClient{},
	)

	// Register query interfaces if needed
	// registry.RegisterImplementations(
	//     (*sdk.Query)(nil),
	//     &QueryClient{},
	// )
}

// RegisterLegacyAminoCodec registers legacy amino codec for backwards compatibility
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// Register message types for amino encoding
	cdc.RegisterConcrete(&MsgRegisterClient{}, "clientregistry/MsgRegisterClient", nil)
	cdc.RegisterConcrete(&MsgStoreAnalysis{}, "clientregistry/MsgStoreAnalysis", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "clientregistry/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgDeactivateClient{}, "clientregistry/MsgDeactivateClient", nil)
}

// GetMarshaler returns the protobuf marshaler
func (c *Codec) GetMarshaler() codec.Codec {
	return c.marshaler
}

// GetLegacyAmino returns the legacy amino codec
func (c *Codec) GetLegacyAmino() *codec.LegacyAmino {
	return c.legacyAmino
}

// GetInterfaceRegistry returns the interface registry
func (c *Codec) GetInterfaceRegistry() types.InterfaceRegistry {
	return c.interfaceRegistry
}

// GetAddressCodec returns the address codec
func (c *Codec) GetAddressCodec() AddressCodec {
	return c.addressCodec
}

// MarshalJSON marshals any value to JSON using protobuf codec
func (c *Codec) MarshalJSON(v interface{}) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.MarshalJSON(msg)
	}
	
	// Fallback to standard JSON marshaling
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals JSON to any value using protobuf codec
func (c *Codec) UnmarshalJSON(data []byte, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.UnmarshalJSON(data, msg)
	}
	
	// Fallback to standard JSON unmarshaling
	return json.Unmarshal(data, v)
}

// MarshalBinary marshals any value to binary using protobuf codec
func (c *Codec) MarshalBinary(v interface{}) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.Marshal(msg)
	}
	
	// Fallback to JSON for non-protobuf messages
	return json.Marshal(v)
}

// UnmarshalBinary unmarshals binary to any value using protobuf codec
func (c *Codec) UnmarshalBinary(data []byte, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.Unmarshal(data, msg)
	}
	
	// Fallback to JSON for non-protobuf messages
	return json.Unmarshal(data, v)
}

// MarshalLengthPrefixed marshals with length prefix
func (c *Codec) MarshalLengthPrefixed(v interface{}) ([]byte, error) {
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.MarshalLengthPrefixed(msg)
	}
	
	// For non-protobuf, marshal to JSON and add length prefix
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	
	// Simple length prefix implementation
	length := len(data)
	result := make([]byte, 4+length)
	result[0] = byte(length >> 24)
	result[1] = byte(length >> 16)
	result[2] = byte(length >> 8)
	result[3] = byte(length)
	copy(result[4:], data)
	
	return result, nil
}

// UnmarshalLengthPrefixed unmarshals with length prefix
func (c *Codec) UnmarshalLengthPrefixed(data []byte, v interface{}) error {
	if len(data) < 4 {
		return fmt.Errorf("data too short for length prefix")
	}
	
	if msg, ok := v.(proto.Message); ok {
		return c.marshaler.UnmarshalLengthPrefixed(data, msg)
	}
	
	// Simple length prefix parsing
	length := int(data[0])<<24 | int(data[1])<<16 | int(data[2])<<8 | int(data[3])
	if len(data) < 4+length {
		return fmt.Errorf("data length mismatch")
	}
	
	return json.Unmarshal(data[4:4+length], v)
}

// ValidateMessage validates a message using the codec
func (c *Codec) ValidateMessage(msg interface{}) error {
	if sdkMsg, ok := msg.(sdk.Msg); ok {
		return sdkMsg.ValidateBasic()
	}
	
	return nil
}

// ValidateAndMarshal validates a message and marshals it
func (c *Codec) ValidateAndMarshal(msg interface{}) ([]byte, error) {
	if err := c.ValidateMessage(msg); err != nil {
		return nil, fmt.Errorf("message validation failed: %w", err)
	}
	
	return c.MarshalBinary(msg)
}

// GetTypeURL returns the type URL for a protobuf message
func (c *Codec) GetTypeURL(msg interface{}) string {
	if protoMsg, ok := msg.(proto.Message); ok {
		msgName := proto.MessageName(protoMsg)
		return "/" + string(msgName)
	}
	
	// Fallback for non-protobuf messages
	return fmt.Sprintf("/%T", msg)
}

// CodecConfig represents codec configuration
type CodecConfig struct {
	Bech32Prefix      string `json:"bech32_prefix"`
	UseProtobuf       bool   `json:"use_protobuf"`
	UseLegacyAmino    bool   `json:"use_legacy_amino"`
	ValidateMessages  bool   `json:"validate_messages"`
}

// DefaultCodecConfig returns default codec configuration
func DefaultCodecConfig() CodecConfig {
	return CodecConfig{
		Bech32Prefix:     "medas",
		UseProtobuf:      true,
		UseLegacyAmino:   true,
		ValidateMessages: true,
	}
}

// JSONCodecConfig returns configuration for JSON codec
func JSONCodecConfig() CodecConfig {
	return CodecConfig{
		Bech32Prefix:     "medas",
		UseProtobuf:      true,
		UseLegacyAmino:   false,
		ValidateMessages: true,
	}
}

// BinaryCodecConfig returns configuration for binary codec
func BinaryCodecConfig() CodecConfig {
	return CodecConfig{
		Bech32Prefix:     "medas",
		UseProtobuf:      true,
		UseLegacyAmino:   false,
		ValidateMessages: false,
	}
}

// NewCodecWithConfig creates a codec with specific configuration
func NewCodecWithConfig(config CodecConfig) *Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewProtoCodec(interfaceRegistry)
	addressCodec := address.NewBech32Codec(config.Bech32Prefix)

	var legacyAmino *codec.LegacyAmino
	if config.UseLegacyAmino {
		legacyAmino = codec.NewLegacyAmino()
		RegisterLegacyAminoCodec(legacyAmino)
	}

	if config.UseProtobuf {
		RegisterInterfaces(interfaceRegistry)
	}

	return &Codec{
		marshaler:         marshaler,
		legacyAmino:       legacyAmino,
		interfaceRegistry: interfaceRegistry,
		addressCodec:      addressCodec,
	}
}

// EncodeAddress encodes an address to string using the address codec
func (c *Codec) EncodeAddress(addr []byte) (string, error) {
	return c.addressCodec.BytesToString(addr)
}

// DecodeAddress decodes an address string to bytes using the address codec
func (c *Codec) DecodeAddress(addr string) ([]byte, error) {
	return c.addressCodec.StringToBytes(addr)
}

// ValidateAddress validates an address format
func (c *Codec) ValidateAddress(addr string) error {
	_, err := c.addressCodec.StringToBytes(addr)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	return nil
}

// NewAddressCodec creates a new address codec with the given prefix
func NewAddressCodec(prefix string) AddressCodec {
	return address.NewBech32Codec(prefix)
}

// MessageTypeURLs returns type URLs for all registered message types
func (c *Codec) MessageTypeURLs() []string {
	return []string{
		"/clientregistry.MsgRegisterClient",
		"/clientregistry.MsgStoreAnalysis", 
		"/clientregistry.MsgUpdateClient",
		"/clientregistry.MsgDeactivateClient",
	}
}

// IsRegisteredMessage checks if a message type is registered
func (c *Codec) IsRegisteredMessage(typeURL string) bool {
	for _, registeredURL := range c.MessageTypeURLs() {
		if typeURL == registeredURL {
			return true
		}
	}
	return false
}

// GetCodecInfo returns information about the codec
func (c *Codec) GetCodecInfo() map[string]interface{} {
	return map[string]interface{}{
		"has_protobuf":      c.marshaler != nil,
		"has_legacy_amino":  c.legacyAmino != nil,
		"address_prefix":    "medas", // Could extract from codec if needed
		"registered_types":  len(c.MessageTypeURLs()),
		"type_urls":         c.MessageTypeURLs(),
	}
}

// Clone creates a copy of the codec
func (c *Codec) Clone() *Codec {
	return &Codec{
		marshaler:         c.marshaler,
		legacyAmino:       c.legacyAmino,
		interfaceRegistry: c.interfaceRegistry,
		addressCodec:      c.addressCodec,
	}
}

// SetAddressCodec sets a new address codec
func (c *Codec) SetAddressCodec(codec AddressCodec) {
	c.addressCodec = codec
}

// Global codec instance for convenience
var GlobalCodec *Codec

// InitGlobalCodec initializes the global codec instance
func InitGlobalCodec() {
	GlobalCodec = NewCodec()
}

// GetGlobalCodec returns the global codec instance
func GetGlobalCodec() *Codec {
	if GlobalCodec == nil {
		InitGlobalCodec()
	}
	return GlobalCodec
}

// Helper functions for common operations

// MustMarshalJSON marshals to JSON and panics on error
func MustMarshalJSON(codec *Codec, v interface{}) []byte {
	bz, err := codec.MarshalJSON(v)
	if err != nil {
		panic(err)
	}
	return bz
}

// MustUnmarshalJSON unmarshals from JSON and panics on error
func MustUnmarshalJSON(codec *Codec, data []byte, v interface{}) {
	err := codec.UnmarshalJSON(data, v)
	if err != nil {
		panic(err)
	}
}

// MustMarshalBinary marshals to binary and panics on error
func MustMarshalBinary(codec *Codec, v interface{}) []byte {
	bz, err := codec.MarshalBinary(v)
	if err != nil {
		panic(err)
	}
	return bz
}

// MustUnmarshalBinary unmarshals from binary and panics on error
func MustUnmarshalBinary(codec *Codec, data []byte, v interface{}) {
	err := codec.UnmarshalBinary(data, v)
	if err != nil {
		panic(err)
	}
}
