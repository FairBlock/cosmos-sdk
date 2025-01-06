package gov

import (
	"fmt"

	kstypes "github.com/Fairblock/fairyring/x/keyshare/types"

	keeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	types "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserror "github.com/cosmos/cosmos-sdk/types/errors"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

type IBCModule struct {
	keeper *keeper.Keeper
}

func NewIBCModule(k *keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	// Require portID is the portID module is bound to
	boundPort := im.keeper.GetPort(ctx)
	if boundPort != portID {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	if version != types.Version {
		return "", sdkerrors.Wrapf(kstypes.ErrInvalidVersion, "got %s, expected %s", version, types.Version)
	}

	// Claim channel capability passed back by IBC module
	if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
		return "", err
	}

	return version, nil
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	// Require portID is the portID module is bound to
	boundPort := im.keeper.GetPort(ctx)
	if boundPort != portID {
		return "", sdkerrors.Wrapf(porttypes.ErrInvalidPort, "invalid port: %s, expected %s", portID, boundPort)
	}

	if counterpartyVersion != types.Version {
		return "", sdkerrors.Wrapf(kstypes.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s", counterpartyVersion, types.Version)
	}

	// Module may have already claimed capability in OnChanOpenInit in the case of crossing hellos
	// (ie chainA and chainB both call ChanOpenInit before one of them calls ChanOpenTry)
	// If module can already authenticate the capability then module already owns it so we don't need to claim
	// Otherwise, module does not have channel capability and we must claim it from IBC
	if !im.keeper.AuthenticateCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)) {
		// Only claim channel capability passed back by IBC module if we do not already own it
		if err := im.keeper.ClaimCapability(ctx, chanCap, host.ChannelCapabilityPath(portID, channelID)); err != nil {
			return "", err
		}
	}
	return types.Version, nil
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_,
	counterpartyVersion string,
) error {
	if counterpartyVersion != types.Version {
		return sdkerrors.Wrapf(kstypes.ErrInvalidVersion, "invalid counterparty version: %s, expected %s", counterpartyVersion, types.Version)
	}
	return nil
}

// OnChanOpenConfirm implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// im.keeper.SetChannel(ctx, channelID)
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for channels
	return sdkerrors.Wrap(cosmoserror.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	fmt.Println("\n\n\n\nReceived Packet\n\n\n\n")

	var ack channeltypes.Acknowledgement

	var modulePacketData kstypes.KeysharePacketData
	if err := kstypes.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &modulePacketData); err != nil {
		return channeltypes.NewErrorAcknowledgement(sdkerrors.Wrapf(cosmoserror.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error()))
	}

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {

	case *kstypes.KeysharePacketData_DecryptionKeyDataPacket:
		fmt.Println("\n\n\n\nReceived Decryption Key Packet\n\n\n\n")
		packetAck, err := im.keeper.OnRecvDecryptionKeyDataPacket(
			ctx, modulePacket,
			*packet.DecryptionKeyDataPacket,
		)
		if err != nil {
			fmt.Println("\n\n\n\nError Processing Packet: ", err, "\n\n\n\n")
			ack = channeltypes.NewErrorAcknowledgement(err)
		} else {
			// Encode packet acknowledgment
			packetAckBytes := v1.MustProtoMarshalJSON(&packetAck)
			ack = channeltypes.NewResultAcknowledgement(sdk.MustSortJSON(packetAckBytes))
		}
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				kstypes.EventTypeDecryptionKeyDataPacket,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(kstypes.AttributeKeyAckSuccess, fmt.Sprintf("%t", err != nil)),
			),
		)
	// this line is used by starport scaffolding # ibc/packet/module/recv
	default:
		err := fmt.Errorf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return channeltypes.NewErrorAcknowledgement(err)
	}

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	fmt.Println("\n\n\n\nReceived Ack Packet\n\n\n\n")

	var ack channeltypes.Acknowledgement
	if err := kstypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(cosmoserror.ErrUnknownRequest, "cannot unmarshal packet acknowledgement: %v", err)
	}

	// this line is used by starport scaffolding # oracle/packet/module/ack

	var modulePacketData kstypes.KeysharePacketData
	if err := kstypes.ModuleCdc.UnmarshalJSON(modulePacket.GetData(), &modulePacketData); err != nil {
		fmt.Println("\n\n\n\nCan't unmarshal Ack Packet\n\n\n\n")
		return sdkerrors.Wrapf(cosmoserror.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	var eventType string

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {
	case *kstypes.KeysharePacketData_RequestDecryptionKeyPacket:
		err := im.keeper.OnAcknowledgementRequestDecryptionKeyPacket(
			ctx,
			modulePacket,
			*packet.RequestDecryptionKeyPacket,
			ack,
		)
		if err != nil {
			return err
		}
		eventType = kstypes.EventTypeRequestDecryptionKeyPacket
	case *kstypes.KeysharePacketData_GetDecryptionKeyPacket:
		err := im.keeper.OnAcknowledgementGetDecryptionKeyPacket(
			ctx,
			modulePacket,
			*packet.GetDecryptionKeyPacket,
			ack,
		)
		if err != nil {
			fmt.Println("\n\n\n Error evaluating OnAcknowledgementGetDecryptionKeyPacket", err, "\n\n\n")
			return err
		}
		eventType = kstypes.EventTypeGetDecryptionKeyPacket

	// this line is used by starport scaffolding # ibc/packet/module/ack
	default:
		errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return sdkerrors.Wrap(cosmoserror.ErrUnknownRequest, errMsg)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			eventType,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(kstypes.AttributeKeyAck, fmt.Sprintf("%v", ack)),
		),
	)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				eventType,
				sdk.NewAttribute(kstypes.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				eventType,
				sdk.NewAttribute(kstypes.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Context,
	modulePacket channeltypes.Packet,
	relayer sdk.AccAddress,
) error {
	fmt.Println("\n\n\n\nReceived timeout Packet\n\n\n\n")

	var modulePacketData kstypes.KeysharePacketData
	if err := modulePacketData.Unmarshal(modulePacket.GetData()); err != nil {
		return sdkerrors.Wrapf(cosmoserror.ErrUnknownRequest, "cannot unmarshal packet data: %s", err.Error())
	}

	// Dispatch packet
	switch packet := modulePacketData.Packet.(type) {

	case *kstypes.KeysharePacketData_RequestDecryptionKeyPacket:
		err := im.keeper.OnTimeoutRequestDecryptionKeyPacket(
			ctx,
			modulePacket,
			*packet.RequestDecryptionKeyPacket,
		)
		if err != nil {
			return err
		}

	case *kstypes.KeysharePacketData_GetDecryptionKeyPacket:
		err := im.keeper.OnTimeoutGetDecryptionKeyPacket(
			ctx,
			modulePacket,
			*packet.GetDecryptionKeyPacket,
		)
		if err != nil {
			return err
		}

	case *kstypes.KeysharePacketData_DecryptionKeyDataPacket:
		err := im.keeper.OnTimeoutDecryptionKeyDataPacket(
			ctx,
			modulePacket,
			*packet.DecryptionKeyDataPacket,
		)
		if err != nil {
			return err
		}
		// this line is used by starport scaffolding # ibc/packet/module/timeout
	default:
		errMsg := fmt.Sprintf("unrecognized %s packet type: %T", types.ModuleName, packet)
		return sdkerrors.Wrap(cosmoserror.ErrUnknownRequest, errMsg)
	}

	return nil
}
