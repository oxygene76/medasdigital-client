package blockchain

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// Module name and route
const (
	ModuleName = "clientregistry"
	RouterKey  = ModuleName
)

// Message type URLs
const (
	TypeMsgRegisterClient = "register_client"
	TypeMsgStoreAnalysis  = "store_analysis"
	TypeMsgUpdateClient   = "update_client"
	TypeMsgDeactivateClient = "deactivate_client"
)

// Ensure messages implement sdk.Msg interface
var (
	_ sdk.Msg = &MsgRegisterClient{}
	_ sdk.Msg = &MsgStoreAnalysis{}
	_ sdk.Msg = &MsgUpdateClient{}
	_ sdk.Msg = &MsgDeactivateClient{}
)

// MsgRegisterClient defines the message for registering a new analysis client
type MsgRegisterClient struct {
	Creator      string   `json:"creator" yaml:"creator"`
	Capabilities []string `json:"capabilities" yaml:"capabilities"`
	Metadata     string   `json:"metadata" yaml:"metadata"`
	GPUInfo      string   `json:"gpu_info,omitempty" yaml:"gpu_info,omitempty"`
}

// NewMsgRegisterClient creates a new MsgRegisterClient
func NewMsgRegisterClient(creator string, capabilities []string, metadata string, gpuInfo string) *MsgRegisterClient {
	return &MsgRegisterClient{
		Creator:      creator,
		Capabilities: capabilities,
		Metadata:     metadata,
		GPUInfo:      gpuInfo,
	}
}

// Route returns the module route
func (msg MsgRegisterClient) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg MsgRegisterClient) Type() string {
	return TypeMsgRegisterClient
}

// GetSigners returns the signers of the message
func (msg MsgRegisterClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the bytes for signing
func (msg MsgRegisterClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgRegisterClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if len(msg.Capabilities) == 0 {
		return fmt.Errorf("capabilities cannot be empty")
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
			return fmt.Errorf("invalid capability: %s", cap)
		}
	}

	return nil
}

// MsgStoreAnalysis defines the message for storing analysis results
type MsgStoreAnalysis struct {
	Creator      string                 `json:"creator" yaml:"creator"`
	ClientID     string                 `json:"client_id" yaml:"client_id"`
	AnalysisType string                 `json:"analysis_type" yaml:"analysis_type"`
	Data         map[string]interface{} `json:"data" yaml:"data"`
	BlockHeight  int64                  `json:"block_height,omitempty" yaml:"block_height,omitempty"`
	TxHash       string                 `json:"tx_hash,omitempty" yaml:"tx_hash,omitempty"`
	Metadata     string                 `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NewMsgStoreAnalysis creates a new MsgStoreAnalysis
func NewMsgStoreAnalysis(creator, clientID, analysisType string, data map[string]interface{}, metadata string) *MsgStoreAnalysis {
	return &MsgStoreAnalysis{
		Creator:      creator,
		ClientID:     clientID,
		AnalysisType: analysisType,
		Data:         data,
		Metadata:     metadata,
	}
}

// Route returns the module route
func (msg MsgStoreAnalysis) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg MsgStoreAnalysis) Type() string {
	return TypeMsgStoreAnalysis
}

// GetSigners returns the signers of the message
func (msg MsgStoreAnalysis) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the bytes for signing
func (msg MsgStoreAnalysis) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgStoreAnalysis) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if msg.ClientID == "" {
		return fmt.Errorf("client_id cannot be empty")
	}

	if msg.AnalysisType == "" {
		return fmt.Errorf("analysis_type cannot be empty")
	}

	if msg.Data == nil || len(msg.Data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"orbital_dynamics":       true,
		"photometric_analysis":   true,
		"astrometric_validation": true,
		"clustering_analysis":    true,
		"survey_processing":      true,
		"anomaly_detection":      true,
		"ai_training":           true,
		"gpu_benchmark":         true,
	}

	if !validTypes[msg.AnalysisType] {
		return fmt.Errorf("invalid analysis type: %s", msg.AnalysisType)
	}

	return nil
}

// MsgUpdateClient defines the message for updating client information
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

// Route returns the module route
func (msg MsgUpdateClient) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg MsgUpdateClient) Type() string {
	return TypeMsgUpdateClient
}

// GetSigners returns the signers of the message
func (msg MsgUpdateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the bytes for signing
func (msg MsgUpdateClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgUpdateClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if msg.ClientID == "" {
		return fmt.Errorf("client_id cannot be empty")
	}

	// Validate status if provided
	if msg.Status != "" {
		validStatuses := map[string]bool{
			"active":      true,
			"inactive":    true,
			"maintenance": true,
			"deactivated": true,
		}

		if !validStatuses[msg.Status] {
			return fmt.Errorf("invalid status: %s", msg.Status)
		}
	}

	return nil
}

// MsgDeactivateClient defines the message for deactivating a client
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

// Route returns the module route
func (msg MsgDeactivateClient) Route() string {
	return RouterKey
}

// Type returns the message type
func (msg MsgDeactivateClient) Type() string {
	return TypeMsgDeactivateClient
}

// GetSigners returns the signers of the message
func (msg MsgDeactivateClient) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the bytes for signing
func (msg MsgDeactivateClient) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validates the message
func (msg MsgDeactivateClient) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return fmt.Errorf("invalid creator address: %w", err)
	}

	if msg.ClientID == "" {
		return fmt.Errorf("client_id cannot be empty")
	}

	return nil
}

// Message service registration
func RegisterMsgServer(server msgservice.Server, keeper MsgServer) {
	RegisterMsgServiceServer(server, keeper)
}

// MsgServer interface for handling messages
type MsgServer interface {
	RegisterClient(ctx sdk.Context, msg *MsgRegisterClient) (*MsgRegisterClientResponse, error)
	StoreAnalysis(ctx sdk.Context, msg *MsgStoreAnalysis) (*MsgStoreAnalysisResponse, error)
	UpdateClient(ctx sdk.Context, msg *MsgUpdateClient) (*MsgUpdateClientResponse, error)
	DeactivateClient(ctx sdk.Context, msg *MsgDeactivateClient) (*MsgDeactivateClientResponse, error)
}

// Response types
type MsgRegisterClientResponse struct {
	ClientID string `json:"client_id" yaml:"client_id"`
}

type MsgStoreAnalysisResponse struct {
	AnalysisID string `json:"analysis_id" yaml:"analysis_id"`
}

type MsgUpdateClientResponse struct {
	Success bool `json:"success" yaml:"success"`
}

type MsgDeactivateClientResponse struct {
	Success bool `json:"success" yaml:"success"`
}

// Helper functions for message creation

// CreateRegistrationMessage creates a registration message with proper formatting
func CreateRegistrationMessage(creator string, capabilities []string, metadata map[string]interface{}, gpuInfo map[string]interface{}) (*MsgRegisterClient, error) {
	// Marshal metadata
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Marshal GPU info
	var gpuInfoStr string
	if gpuInfo != nil {
		gpuInfoBytes, err := json.Marshal(gpuInfo)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal GPU info: %w", err)
		}
		gpuInfoStr = string(gpuInfoBytes)
	}

	return NewMsgRegisterClient(creator, capabilities, string(metadataBytes), gpuInfoStr), nil
}

// CreateAnalysisMessage creates an analysis storage message
func CreateAnalysisMessage(creator, clientID, analysisType string, results map[string]interface{}, metadata map[string]interface{}) (*MsgStoreAnalysis, error) {
	// Marshal metadata
	var metadataStr string
	if metadata != nil {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataStr = string(metadataBytes)
	}

	return NewMsgStoreAnalysis(creator, clientID, analysisType, results, metadataStr), nil
}

// ValidateCapabilities validates a list of capabilities
func ValidateCapabilities(capabilities []string) error {
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

	for _, cap := range capabilities {
		if !validCapabilities[cap] {
			return fmt.Errorf("invalid capability: %s", cap)
		}
	}

	return nil
}

// ParseMetadata parses metadata JSON string
func ParseMetadata(metadataStr string) (map[string]interface{}, error) {
	if metadataStr == "" {
		return make(map[string]interface{}), nil
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// FormatAnalysisData formats analysis data for storage
func FormatAnalysisData(data interface{}) (map[string]interface{}, error) {
	// Convert data to map[string]interface{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis data: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis data: %w", err)
	}

	return result, nil
}

// Message codec
var ModuleCdc = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(ModuleCdc)
	ModuleCdc.Seal()
}

// RegisterLegacyAminoCodec registers the necessary interfaces and concrete types
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterClient{}, "clientregistry/MsgRegisterClient", nil)
	cdc.RegisterConcrete(&MsgStoreAnalysis{}, "clientregistry/MsgStoreAnalysis", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "clientregistry/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgDeactivateClient{}, "clientregistry/MsgDeactivateClient", nil)
}
