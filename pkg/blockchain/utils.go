package blockchain

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	
	// Import for protobuf registration (v0.50)
	"github.com/cosmos/gogoproto/proto"
)

// RegisterLegacyAminoCodec registers the necessary x/clientregistry interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization. (Updated for v0.50)
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterClient{}, "clientregistry/MsgRegisterClient", nil)
	cdc.RegisterConcrete(&MsgStoreAnalysis{}, "clientregistry/MsgStoreAnalysis", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "clientregistry/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgDeactivateClient{}, "clientregistry/MsgDeactivateClient", nil)
}

// RegisterInterfaces registers the x/clientregistry interfaces types with the
// interface registry (Updated for v0.50)
func RegisterInterfaces(registry types.InterfaceRegistry) {
	// Register message implementations
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterClient{},
		&MsgStoreAnalysis{},
		&MsgUpdateClient{},
		&MsgDeactivateClient{},
	)

	// Register message service descriptor (v0.50 style)
	// Note: This would normally be auto-generated from protobuf files
	// For now, we'll handle it manually for compatibility
	registerMsgServiceDesc(registry)
}

// registerMsgServiceDesc manually registers the message service descriptor
// This would normally be auto-generated from protobuf in a real v0.50 implementation
func registerMsgServiceDesc(registry types.InterfaceRegistry) {
	// Create a simple service descriptor for our messages
	// In a real implementation, this would come from generated protobuf code
	
	// For v0.50, we need to ensure the messages are properly registered
	// with their full type URLs for protobuf compatibility
	registry.RegisterImplementations(
		(*proto.Message)(nil),
		&MsgRegisterClient{},
		&MsgStoreAnalysis{},
		&MsgUpdateClient{},
		&MsgDeactivateClient{},
	)
}

// NewInterfaceRegistry creates a new interface registry with our types registered
func NewInterfaceRegistry() types.InterfaceRegistry {
	registry := types.NewInterfaceRegistry()
	RegisterInterfaces(registry)
	return registry
}

// NewCodec creates a new codec with all necessary types registered
func NewCodec() codec.Codec {
	registry := NewInterfaceRegistry()
	return codec.NewProtoCodec(registry)
}

// Variables for module-wide codec usage (updated for v0.50)
var (
	// Amino is the legacy amino codec
	Amino = codec.NewLegacyAmino()
	
	// ModuleCdc is the codec for the module (v0.50 style)
	ModuleCdc = NewCodec()
	
	// AddressCodec for address encoding/decoding (new in v0.50)
	AddressCodec = address.NewBech32Codec("medas")
)

func init() {
	// Register legacy amino codec
	RegisterLegacyAminoCodec(Amino)
	Amino.Seal()
}

// Codec interface implementations for v0.50 compatibility

// MarshalJSON marshals a message to JSON using the module codec
func MarshalJSON(msg interface{}) ([]byte, error) {
	return ModuleCdc.MarshalJSON(msg)
}

// UnmarshalJSON unmarshals JSON to a message using the module codec
func UnmarshalJSON(bz []byte, ptr interface{}) error {
	return ModuleCdc.UnmarshalJSON(bz, ptr)
}

// MustMarshalJSON marshals to JSON and panics on error
func MustMarshalJSON(msg interface{}) []byte {
	bz, err := MarshalJSON(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// MustUnmarshalJSON unmarshals JSON and panics on error
func MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}

// MarshalBinary marshals a message to binary using protobuf (new in v0.50)
func MarshalBinary(msg proto.Message) ([]byte, error) {
	return ModuleCdc.Marshal(msg)
}

// UnmarshalBinary unmarshals binary to a message using protobuf (new in v0.50)
func UnmarshalBinary(bz []byte, ptr proto.Message) error {
	return ModuleCdc.Unmarshal(bz, ptr)
}

// MustMarshalBinary marshals to binary and panics on error (new in v0.50)
func MustMarshalBinary(msg proto.Message) []byte {
	bz, err := MarshalBinary(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

// MustUnmarshalBinary unmarshals binary and panics on error (new in v0.50)
func MustUnmarshalBinary(bz []byte, ptr proto.Message) {
	if err := UnmarshalBinary(bz, ptr); err != nil {
		panic(err)
	}
}

// GetModuleCodec returns the module codec
func GetModuleCodec() codec.Codec {
	return ModuleCdc
}

// GetLegacyCodec returns the legacy amino codec
func GetLegacyCodec() *codec.LegacyAmino {
	return Amino
}

// GetAddressCodec returns the address codec (new in v0.50)
func GetAddressCodec() address.Codec {
	return AddressCodec
}

// RegisterCodec registers all necessary types with the given codec
// This function can be used to register types with external codecs
// when needed for integration with other modules
func RegisterCodec(cdc codec.Codec) {
	// This function can be used to register types with external codecs
	// when needed for integration with other modules
	if legacyAmino, ok := cdc.(*codec.LegacyAmino); ok {
		RegisterLegacyAminoCodec(legacyAmino)
	}
}

// RegisterLegacyAminoCodecWithTypes registers types with an external legacy amino codec
func RegisterLegacyAminoCodecWithTypes(cdc *codec.LegacyAmino) {
	RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfacesWithRegistry registers interfaces with an external interface registry
func RegisterInterfacesWithRegistry(registry types.InterfaceRegistry) {
	RegisterInterfaces(registry)
}

// Codec configuration for different encoding formats (v0.50 style)

// JSONCodecConfig returns codec configuration for JSON encoding
type JSONCodecConfig struct {
	Codec codec.Codec
}

// NewJSONCodecConfig creates a new JSON codec configuration
func NewJSONCodecConfig() *JSONCodecConfig {
	return &JSONCodecConfig{
		Codec: ModuleCdc,
	}
}

// BinaryCodecConfig returns codec configuration for binary encoding
type BinaryCodecConfig struct {
	Codec codec.Codec
}

// NewBinaryCodecConfig creates a new binary codec configuration
func NewBinaryCodecConfig() *BinaryCodecConfig {
	return &BinaryCodecConfig{
		Codec: ModuleCdc,
	}
}

// Utility functions for message type URLs (v0.50 style)

// GetTypeURL returns the type URL for a message (used in protobuf)
func GetTypeURL(msg proto.Message) string {
	return "/" + proto.MessageName(msg)
}

// GetMsgTypeURL returns the type URL for our custom messages
func GetMsgTypeURL(msgType string) string {
	switch msgType {
	case TypeMsgRegisterClient:
		return "/clientregistry.MsgRegisterClient"
	case TypeMsgStoreAnalysis:
		return "/clientregistry.MsgStoreAnalysis"
	case TypeMsgUpdateClient:
		return "/clientregistry.MsgUpdateClient"
	case TypeMsgDeactivateClient:
		return "/clientregistry.MsgDeactivateClient"
	default:
		return ""
	}
}

// IsRegisteredMessage checks if a message type is registered
func IsRegisteredMessage(msgType string) bool {
	return GetMsgTypeURL(msgType) != ""
}

// Validation helpers for v0.50

// ValidateMessage validates a message using the codec
func ValidateMessage(msg sdk.Msg) error {
	// First validate using the message's own ValidateBasic method
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	// Additional codec-level validation could go here
	// For example, ensuring the message can be properly marshaled/unmarshaled
	
	return nil
}

// ValidateAndMarshal validates a message and marshals it to JSON
func ValidateAndMarshal(msg sdk.Msg) ([]byte, error) {
	if err := ValidateMessage(msg); err != nil {
		return nil, err
	}
	
	return MarshalJSON(msg)
}

// UnmarshalAndValidate unmarshals a message from JSON and validates it
func UnmarshalAndValidate(bz []byte, msg sdk.Msg) error {
	if err := UnmarshalJSON(bz, msg); err != nil {
		return err
	}
	
	return ValidateMessage(msg)
}
