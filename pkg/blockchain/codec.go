package blockchain

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/clientregistry interfaces
// and concrete types on the provided LegacyAmino codec. These types are used
// for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterClient{}, "clientregistry/MsgRegisterClient", nil)
	cdc.RegisterConcrete(&MsgStoreAnalysis{}, "clientregistry/MsgStoreAnalysis", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "clientregistry/MsgUpdateClient", nil)
}

// RegisterInterfaces registers the x/clientregistry interfaces types with the
// interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterClient{},
		&MsgStoreAnalysis{},
		&MsgUpdateClient{},
	)

	// Note: This would normally be generated from protobuf
	// For now we'll comment it out until the proto files are generated
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// Amino is the legacy amino codec
	Amino = codec.NewLegacyAmino()
	
	// ModuleCdc is the codec for the module
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(Amino)
	Amino.Seal()
}

// Custom marshal/unmarshal functions for compatibility

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

// GetModuleCodec returns the module codec
func GetModuleCodec() codec.Codec {
	return ModuleCdc
}

// RegisterCodec registers all necessary types with the given codec
func RegisterCodec(cdc codec.Codec) {
	// This function can be used to register types with external codecs
	// when needed for integration with other modules
}
