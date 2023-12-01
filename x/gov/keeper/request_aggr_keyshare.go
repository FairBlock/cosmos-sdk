package keeper

import (
	"errors"
	"fmt"
	"strconv"

	kstypes "fairyring/x/keyshare/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
)

// TransmitRequestAggrKeysharePacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitRequestAggrKeysharePacket(
	ctx sdk.Context,
	packetData kstypes.RequestAggrKeysharePacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	fmt.Println("\n\n\nTransmitRequestAggrKeysharePacket\n\n\n")

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

// OnAcknowledgementRequestAggrKeysharePacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementRequestAggrKeysharePacket(ctx sdk.Context, packet channeltypes.Packet, data kstypes.RequestAggrKeysharePacketData, ack channeltypes.Acknowledgement) error {
	fmt.Println("\n\n\nOnAcknowledgementRequestAggrKeysharePacket\n\n\n")
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:

		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error
		fmt.Println("\n\n\nOnAcknowledgementRequestAggrKeysharePacket failure for reqID: ", data.ProposalId)

		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck kstypes.RequestAggrKeysharePacketAck

		if err := kstypes.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		pID, _ := strconv.ParseUint(data.ProposalId, 10, 64)
		proposal, found := k.GetProposal(ctx, pID)
		if !found {
			return errors.New("Proposal not found")
		}

		proposal.Identity = packetAck.Identity
		proposal.Pubkey = packetAck.Pubkey

		k.SetProposal(ctx, proposal)
		fmt.Println("\n\n\nOnAcknowledgementRequestAggrKeysharePacket success for reqID: ", data.ProposalId)

		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutRequestAggrKeysharePacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutRequestAggrKeysharePacket(ctx sdk.Context, packet channeltypes.Packet, data kstypes.RequestAggrKeysharePacketData) error {

	// TODO: packet timeout logic

	return nil
}
