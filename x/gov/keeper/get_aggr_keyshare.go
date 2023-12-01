package keeper

import (
	"errors"
	"fmt"

	"fairyring/x/keyshare/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
)

// TransmitGetAggrKeysharePacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitGetAggrKeysharePacket(
	ctx sdk.Context,
	packetData types.GetAggrKeysharePacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	fmt.Println("\n\n\nTransmitGetAggrKeysharePacket\n\n\n")

	channelCap, ok := k.ScopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes := packetData.GetBytes()
	// if err != nil {
	// 	return 0, sdkerrors.Wrapf(sdkerrors.ErrJSONMarshal, "cannot marshal the packet: %w", err)
	// }

	return k.ChannelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetBytes)
}

// OnAcknowledgementGetAggrKeysharePacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementGetAggrKeysharePacket(ctx sdk.Context, packet channeltypes.Packet, data types.GetAggrKeysharePacketData, ack channeltypes.Acknowledgement) error {
	fmt.Println("\n\n\nOnAcknowledgementGetAggrKeysharePacket\n\n\n")

	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:

		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error
		fmt.Println("\n\n\nnOnAcknowledgementGetAggrKeysharePacket failure for reqID: ", data.Identity)

		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.GetAggrKeysharePacketAck

		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		fmt.Println("\n\n\nnOnAcknowledgementGetAggrKeysharePacket success for reqID: ", data.Identity)

		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutGetAggrKeysharePacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutGetAggrKeysharePacket(ctx sdk.Context, packet channeltypes.Packet, data types.GetAggrKeysharePacketData) error {

	// TODO: packet timeout logic

	return nil
}
