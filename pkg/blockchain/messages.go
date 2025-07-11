package blockchain

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// Module name
	ModuleName = "clientregistry"
	
	// Message types
	TypeMsgRegisterClient   = "register_client"
	TypeMsgStoreAnalysis    = "store_analysis"
	TypeMsgUpdateClient     = "update_client"
	TypeMsgDeactivateClient = "deactivate_client"
)

// MsgRegisterClient defines a message for registering a new analysis client
type MsgRegisterClient struct {
	Creator      string   `json:"creator" yaml:"creator"`
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
	Metadata     string   `json:"metadata" yaml:"metadata"`
}

// NewMsgRegisterClient creates a new MsgRegisterClient
func NewMsgRegisterClient(creator string, capabilities []string, metadata string) *MsgRegisterClient {
	return &MsgRegisterClient{
		Creator:      creator,
		Capabilities: capabilities,
		Metadata:     metadata,
	}
}

// Route returns the message route
func (msg MsgRegisterClient) Route() string {
	return ModuleName
}

// Type returns the message type
func (msg MsgRegisterClient) Type() string {
	return TypeMsgRegisterClient
}

// GetSigners returns the required signers
func (msg MsgRegisterClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgRegisterClient) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgRegisterClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.Capabilities) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "capabilities cannot be empty")
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

	for _, cap := range msg.Capabilities {
		if !validCapabilities[cap] {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid capability: %s", cap)
		}
	}

	if len(msg.Metadata) > 10000 { // 10KB limit
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata too large (max 10KB)")
	}

	return nil
}

// MsgStoreAnalysis defines a message for storing analysis results
type MsgStoreAnalysis struct {
	Creator      string                 `json:"creator" yaml:"creator"`
	ClientID     string                 `json:"client_id" yaml:"client_id"`
	AnalysisType string                 `json:"analysis_type" yaml:"analysis_type"`
	Data         map[string]interface{} `json:"data" yaml:"data"`
	BlockHeight  int64                  `json:"block_height,omitempty" yaml:"block_height,omitempty"`
	TxHash       string                 `json:"tx_hash,omitempty" yaml:"tx_hash,omitempty"`
}

// NewMsgStoreAnalysis creates a new MsgStoreAnalysis
func NewMsgStoreAnalysis(creator, clientID, analysisType string, data map[string]interface{}, blockHeight int64, txHash string) *MsgStoreAnalysis {
	return &MsgStoreAnalysis{
		Creator:      creator,
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         data,
		BlockHeight:  blockHeight,
		TxHash:       txHash,
	}
}

// Route returns the message route
func (msg MsgStoreAnalysis) Route() string {
	return ModuleName
}

// Type returns the message type
func (msg MsgStoreAnalysis) Type() string {
	return TypeMsgStoreAnalysis
}

// GetSigners returns the required signers
func (msg MsgStoreAnalysis) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgStoreAnalysis) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgStoreAnalysis) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.ClientID) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "client ID cannot be empty")
	}

	if len(msg.AnalysisType) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "analysis type cannot be empty")
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"orbital_dynamics":      true,
		"photometric_analysis":  true,
		"astrometric_validation": true,
		"clustering_analysis":   true,
		"survey_processing":     true,
		"anomaly_detection":     true,
		"ai_training":          true,
		"ai_detection":         true,
	}

	if !validTypes[msg.AnalysisType] {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid analysis type: %s", msg.AnalysisType)
	}

	if msg.Data == nil || len(msg.Data) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "analysis data cannot be empty")
	}

	// Check data size (serialize to estimate size)
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "invalid analysis data format")
	}

	if len(dataBytes) > 1000000 { // 1MB limit
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "analysis data too large (max 1MB)")
	}

	return nil
}

// MsgUpdateClient defines a message for updating client information
type MsgUpdateClient struct {
	Creator      string   `json:"creator" yaml:"creator"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	Capabilities []string `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
	Metadata     string   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Status       string   `json:"status,omitempty" yaml:"status,omitempty"`
}

// NewMsgUpdateClient creates a new MsgUpdateClient
func NewMsgUpdateClient(creator, clientID string, capabilities []string, metadata, status string) *MsgUpdateClient {
	return &MsgUpdateClient{
		Creator:      creator,
		ClientID:     clientID,
		Capabilities: capabilities,
		Metadata:     metadata,
		Status:       status,
	}
}

// Route returns the message route
func (msg MsgUpdateClient) Route() string {
	return ModuleName
}

// Type returns the message type
func (msg MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// GetSigners returns the required signers
func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgUpdateClient) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgUpdateClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.ClientID) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "client ID cannot be empty")
	}

	// Validate status if provided
	if msg.Status != "" {
		validStatuses := map[string]bool{
			"active":     true,
			"inactive":   true,
			"suspended":  true,
			"maintenance": true,
		}

		if !validStatuses[msg.Status] {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid status: %s", msg.Status)
		}
	}

	return nil
}

// MsgDeactivateClient defines a message for deactivating a client
type MsgDeactivateClient struct {
	Creator  string `json:"creator" yaml:"creator"`
	ClientID string `json:"client_id" yaml:"client_id"`
	Reason   string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// NewMsgDeactivateClient creates a new MsgDeactivateClient
func NewMsgDeactivateClient(creator, clientID, reason string) *MsgDeactivateClient {
	return &MsgDeactivateClient{
		Creator:  creator,
		ClientID: clientID,
		Reason:   reason,
	}
}

// Route returns the message route
func (msg MsgDeactivateClient) Route() string {
	return ModuleName
}

// Type returns the message type
func (msg MsgDeactivateClient) Type() string {
	return TypeMsgDeactivateClient
}

// GetSigners returns the required signers
func (msg MsgDeactivateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over
func (msg MsgDeactivateClient) GetSignBytes() []byte {
	bz, _ := json.Marshal(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgDeactivateClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.ClientID) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "client ID cannot be empty")
	}

	return nil
}

// Event types
const (
	EventTypeRegisterClient   = "register_client"
	EventTypeStoreAnalysis    = "store_analysis"
	EventTypeUpdateClient     = "update_client"
	EventTypeDeactivateClient = "deactivate_client"

	AttributeKeyClientID     = "client_id"
	AttributeKeyCreator      = "creator"
	AttributeKeyCapabilities = "capabilities"
	AttributeKeyAnalysisType = "analysis_type"
	AttributeKeyStatus       = "status"
	AttributeKeyReason       = "reason"
)

// CreateRegisterClientEvent creates an event for client registration
func CreateRegisterClientEvent(clientID, creator string, capabilities []string) sdk.Event {
	capJSON, _ := json.Marshal(capabilities)
	
	return sdk.NewEvent(
		EventTypeRegisterClient,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyCapabilities, string(capJSON)),
	)
}

// CreateStoreAnalysisEvent creates an event for analysis storage
func CreateStoreAnalysisEvent(clientID, creator, analysisType string) sdk.Event {
	return sdk.NewEvent(
		EventTypeStoreAnalysis,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyAnalysisType, analysisType),
	)
}

// CreateUpdateClientEvent creates an event for client update
func CreateUpdateClientEvent(clientID, creator, status string) sdk.Event {
	return sdk.NewEvent(
		EventTypeUpdateClient,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyStatus, status),
	)
}

// CreateDeactivateClientEvent creates an event for client deactivation
func CreateDeactivateClientEvent(clientID, creator, reason string) sdk.Event {
	return sdk.NewEvent(
		EventTypeDeactivateClient,
		sdk.NewAttribute(AttributeKeyClientID, clientID),
		sdk.NewAttribute(AttributeKeyCreator, creator),
		sdk.NewAttribute(AttributeKeyReason, reason),
	)
}
