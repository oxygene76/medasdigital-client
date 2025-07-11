package blockchain

import (
	"fmt"
	"strings"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// Ensure all messages implement the required interfaces
var (
	_ sdk.Msg       = (*MsgRegisterClient)(nil)
	_ proto.Message = (*MsgRegisterClient)(nil)
	_ sdk.Msg       = (*MsgStoreAnalysis)(nil)
	_ proto.Message = (*MsgStoreAnalysis)(nil)
	_ sdk.Msg       = (*MsgUpdateClient)(nil)
	_ proto.Message = (*MsgUpdateClient)(nil)
	_ sdk.Msg       = (*MsgDeactivateClient)(nil)
	_ proto.Message = (*MsgDeactivateClient)(nil)
)

// Error definitions
var (
	ErrInvalidMessage     = errors.Register("blockchain", 1, "invalid message")
	ErrInvalidAddress     = errors.Register("blockchain", 2, "invalid address")
	ErrInvalidCapability  = errors.Register("blockchain", 3, "invalid capability")
	ErrInvalidAnalysisType = errors.Register("blockchain", 4, "invalid analysis type")
)

// Message type constants
const (
	TypeMsgRegisterClient   = "register_client"
	TypeMsgStoreAnalysis    = "store_analysis"
	TypeMsgUpdateClient     = "update_client"
	TypeMsgDeactivateClient = "deactivate_client"
)

// Route constants
const (
	ModuleName = "clientregistry"
)

// MsgRegisterClient defines the message for registering a new client
type MsgRegisterClient struct {
	Creator      string   `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Capabilities []string `protobuf:"bytes,2,rep,name=capabilities,proto3" json:"capabilities,omitempty"`
	Metadata     string   `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

// Route implements sdk.Msg interface (legacy)
func (msg *MsgRegisterClient) Route() string {
	return ModuleName
}

// Type implements sdk.Msg interface (legacy)
func (msg *MsgRegisterClient) Type() string {
	return TypeMsgRegisterClient
}

// GetSigners implements sdk.Msg interface
func (msg *MsgRegisterClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignersStr returns signers as strings (v0.50 requirement)
func (msg *MsgRegisterClient) GetSignersStr() ([]string, error) {
	return []string{msg.Creator}, nil
}

// GetSignBytes implements sdk.Msg interface (legacy)
func (msg *MsgRegisterClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation of the message
func (msg *MsgRegisterClient) ValidateBasic() error {
	if msg.Creator == "" {
		return errors.Wrap(ErrInvalidMessage, "creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errors.Wrapf(ErrInvalidMessage, "invalid creator address: %v", err)
	}

	if len(msg.Capabilities) == 0 {
		return errors.Wrap(ErrInvalidMessage, "capabilities cannot be empty")
	}

	// Validate capabilities
	validCapabilities := map[string]bool{
		"orbital_dynamics":       true,
		"photometric_analysis":   true,
		"astrometric_validation": true,
		"clustering_analysis":    true,
		"survey_processing":      true,
		"anomaly_detection":      true,
		"ai_training":           true,
		"gpu_compute":           true,
	}

	for _, capability := range msg.Capabilities {
		if !validCapabilities[capability] {
			return errors.Wrapf(ErrInvalidMessage, "invalid capability: %s", capability)
		}
	}

	// Validate metadata length
	if len(msg.Metadata) > 10*1024 { // 10KB limit
		return errors.Wrap(ErrInvalidMessage, "metadata too large (max 10KB)")
	}

	return nil
}

// ProtoMessage implements proto.Message interface
func (msg *MsgRegisterClient) ProtoMessage() {}

// Reset implements proto.Message interface
func (msg *MsgRegisterClient) Reset() {
	*msg = MsgRegisterClient{}
}

// String implements proto.Message interface
func (msg *MsgRegisterClient) String() string {
	return fmt.Sprintf("MsgRegisterClient{Creator: %s, Capabilities: %v}", msg.Creator, msg.Capabilities)
}

// MsgStoreAnalysis defines the message for storing analysis results
type MsgStoreAnalysis struct {
	Creator      string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	ClientID     string `protobuf:"bytes,2,opt,name=client_id,proto3" json:"client_id,omitempty"`
	AnalysisType string `protobuf:"bytes,3,opt,name=analysis_type,proto3" json:"analysis_type,omitempty"`
	Data         string `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
	BlockHeight  int64  `protobuf:"varint,5,opt,name=block_height,proto3" json:"block_height,omitempty"`
	TxHash       string `protobuf:"bytes,6,opt,name=tx_hash,proto3" json:"tx_hash,omitempty"`
}

// Route implements sdk.Msg interface (legacy)
func (msg *MsgStoreAnalysis) Route() string {
	return ModuleName
}

// Type implements sdk.Msg interface (legacy)
func (msg *MsgStoreAnalysis) Type() string {
	return TypeMsgStoreAnalysis
}

// GetSigners implements sdk.Msg interface
func (msg *MsgStoreAnalysis) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignersStr returns signers as strings (v0.50 requirement)
func (msg *MsgStoreAnalysis) GetSignersStr() ([]string, error) {
	return []string{msg.Creator}, nil
}

// GetSignBytes implements sdk.Msg interface (legacy)
func (msg *MsgStoreAnalysis) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation of the message
func (msg *MsgStoreAnalysis) ValidateBasic() error {
	if msg.Creator == "" {
		return errors.Wrap(ErrInvalidMessage, "creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errors.Wrapf(ErrInvalidMessage, "invalid creator address: %v", err)
	}

	if msg.ClientID == "" {
		return errors.Wrap(ErrInvalidMessage, "client_id cannot be empty")
	}

	if msg.AnalysisType == "" {
		return errors.Wrap(ErrInvalidMessage, "analysis_type cannot be empty")
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"orbital_dynamics":  true,
		"photometric":       true,
		"clustering":        true,
		"ai_training":       true,
		"anomaly_detection": true,
		"survey_processing": true,
	}

	if !validTypes[msg.AnalysisType] {
		return errors.Wrapf(ErrInvalidMessage, "invalid analysis type: %s", msg.AnalysisType)
	}

	// Validate data length
	if len(msg.Data) > 1024*1024 { // 1MB limit
		return errors.Wrap(ErrInvalidMessage, "analysis data too large (max 1MB)")
	}

	return nil
}

// ProtoMessage implements proto.Message interface
func (msg *MsgStoreAnalysis) ProtoMessage() {}

// Reset implements proto.Message interface
func (msg *MsgStoreAnalysis) Reset() {
	*msg = MsgStoreAnalysis{}
}

// String implements proto.Message interface
func (msg *MsgStoreAnalysis) String() string {
	return fmt.Sprintf("MsgStoreAnalysis{Creator: %s, ClientID: %s, Type: %s}", msg.Creator, msg.ClientID, msg.AnalysisType)
}

// MsgUpdateClient defines the message for updating client capabilities
type MsgUpdateClient struct {
	Creator         string   `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	ClientID        string   `protobuf:"bytes,2,opt,name=client_id,proto3" json:"client_id,omitempty"`
	NewCapabilities []string `protobuf:"bytes,3,rep,name=new_capabilities,proto3" json:"new_capabilities,omitempty"`
	NewMetadata     string   `protobuf:"bytes,4,opt,name=new_metadata,proto3" json:"new_metadata,omitempty"`
}

// Route implements sdk.Msg interface (legacy)
func (msg *MsgUpdateClient) Route() string {
	return ModuleName
}

// Type implements sdk.Msg interface (legacy)
func (msg *MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// GetSigners implements sdk.Msg interface
func (msg *MsgUpdateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignersStr returns signers as strings (v0.50 requirement)
func (msg *MsgUpdateClient) GetSignersStr() ([]string, error) {
	return []string{msg.Creator}, nil
}

// GetSignBytes implements sdk.Msg interface (legacy)
func (msg *MsgUpdateClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation of the message
func (msg *MsgUpdateClient) ValidateBasic() error {
	if msg.Creator == "" {
		return errors.Wrap(ErrInvalidMessage, "creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errors.Wrapf(ErrInvalidMessage, "invalid creator address: %v", err)
	}

	if msg.ClientID == "" {
		return errors.Wrap(ErrInvalidMessage, "client_id cannot be empty")
	}

	if len(msg.NewCapabilities) == 0 {
		return errors.Wrap(ErrInvalidMessage, "new_capabilities cannot be empty")
	}

	return nil
}

// ProtoMessage implements proto.Message interface
func (msg *MsgUpdateClient) ProtoMessage() {}

// Reset implements proto.Message interface
func (msg *MsgUpdateClient) Reset() {
	*msg = MsgUpdateClient{}
}

// String implements proto.Message interface
func (msg *MsgUpdateClient) String() string {
	return fmt.Sprintf("MsgUpdateClient{Creator: %s, ClientID: %s}", msg.Creator, msg.ClientID)
}

// MsgDeactivateClient defines the message for deactivating a client
type MsgDeactivateClient struct {
	Creator  string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	ClientID string `protobuf:"bytes,2,opt,name=client_id,proto3" json:"client_id,omitempty"`
	Reason   string `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
}

// Route implements sdk.Msg interface (legacy)
func (msg *MsgDeactivateClient) Route() string {
	return ModuleName
}

// Type implements sdk.Msg interface (legacy)
func (msg *MsgDeactivateClient) Type() string {
	return TypeMsgDeactivateClient
}

// GetSigners implements sdk.Msg interface
func (msg *MsgDeactivateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignersStr returns signers as strings (v0.50 requirement)
func (msg *MsgDeactivateClient) GetSignersStr() ([]string, error) {
	return []string{msg.Creator}, nil
}

// GetSignBytes implements sdk.Msg interface (legacy)
func (msg *MsgDeactivateClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic performs basic validation of the message
func (msg *MsgDeactivateClient) ValidateBasic() error {
	if msg.Creator == "" {
		return errors.Wrap(ErrInvalidMessage, "creator cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return errors.Wrapf(ErrInvalidMessage, "invalid creator address: %v", err)
	}

	if msg.ClientID == "" {
		return errors.Wrap(ErrInvalidMessage, "client_id cannot be empty")
	}

	return nil
}

// ProtoMessage implements proto.Message interface
func (msg *MsgDeactivateClient) ProtoMessage() {}

// Reset implements proto.Message interface
func (msg *MsgDeactivateClient) Reset() {
	*msg = MsgDeactivateClient{}
}

// String implements proto.Message interface
func (msg *MsgDeactivateClient) String() string {
	return fmt.Sprintf("MsgDeactivateClient{Creator: %s, ClientID: %s}", msg.Creator, msg.ClientID)
}

// Event type constants
const (
	EventTypeClientRegistered   = "client_registered"
	EventTypeAnalysisStored     = "analysis_stored"
	EventTypeClientUpdated      = "client_updated"
	EventTypeClientDeactivated  = "client_deactivated"

	AttributeKeyClientID       = "client_id"
	AttributeKeyCreator        = "creator"
	AttributeKeyCapabilities   = "capabilities"
	AttributeKeyAnalysisType   = "analysis_type"
	AttributeKeyBlockHeight    = "block_height"
	AttributeKeyTxHash         = "tx_hash"
	AttributeKeyStatus         = "status"
	AttributeKeyReason         = "reason"
)

// NewClientRegisteredEvent creates a new client registered event
func NewClientRegisteredEvent(clientID, creator string, capabilities []string) sdk.Event {
	return sdk.NewEvent(
		EventTypeClientRegistered,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyCapabilities, strings.Join(capabilities, ",")),
		sdk.NewAttribute(AttributeKeyStatus, "active"),
	)
}

// NewAnalysisStoredEvent creates a new analysis stored event
func NewAnalysisStoredEvent(clientID, creator, analysisType, txHash string, blockHeight int64) sdk.Event {
	return sdk.NewEvent(
		EventTypeAnalysisStored,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyAnalysisType, analysisType),
		sdk.NewAttribute(AttributeKeyBlockHeight, fmt.Sprintf("%d", blockHeight)),
		sdk.NewAttribute(AttributeKeyTxHash, txHash),
		sdk.NewAttribute(AttributeKeyStatus, "stored"),
	)
}

// NewClientUpdatedEvent creates a new client updated event
func NewClientUpdatedEvent(clientID, creator string, newCapabilities []string) sdk.Event {
	return sdk.NewEvent(
		EventTypeClientUpdated,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyCapabilities, strings.Join(newCapabilities, ",")),
		sdk.NewAttribute(AttributeKeyStatus, "updated"),
	)
}

// NewClientDeactivatedEvent creates a new client deactivated event
func NewClientDeactivatedEvent(clientID, creator, reason string) sdk.Event {
	return sdk.NewEvent(
		EventTypeClientDeactivated,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyReason, reason),
		sdk.NewAttribute(AttributeKeyStatus, "deactivated"),
	)
}

// Legacy Amino codec (will be set by codec.go)
var ModuleCdc interface {
	MustMarshalJSON(o interface{}) []byte
}
