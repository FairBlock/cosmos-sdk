package keeper

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	kstypes "github.com/Fairblock/fairyring/x/keyshare/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// TransmitRequestDecryptionKeyPacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitRequestDecryptionKeyPacket(
	ctx sdk.Context,
	packetData kstypes.RequestDecryptionKeyPacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	channelCap, ok := k.ScopedKeeper().GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes := packetData.GetBytes()
	// if err != nil {
	// 	return 0, sdkerrors.Wrapf(sdkerrors.ErrJSONMarshal, "cannot marshal the packet: %w", err)
	// }

	return k.ibcKeeperFn().ChannelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetBytes)
}

// OnAcknowledgementRequestDecryptionKeyPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementRequestDecryptionKeyPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data kstypes.RequestDecryptionKeyPacketData,
	ack channeltypes.Acknowledgement,
) error {
	fmt.Println("\n\n\n\nOnAcknowledgementRequestDecryptionKeyPacket\n\n\n\n")
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:

		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck kstypes.RequestDecryptionKeyPacketAck

		if err := kstypes.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		pID, _ := strconv.ParseUint(data.GetProposalId(), 10, 64)
		proposal, found := k.GetProposal(ctx, pID)
		if !found {
			return errors.New("Proposal not found")
		}

		proposal.Identity = packetAck.Identity
		proposal.Pubkey = packetAck.Pubkey

		fmt.Println("\n\n\n\nOnAcknowledgementRequestDecryptionKeyPacket Acknowledgement_Result\n\n\n\n")

		k.SetProposal(ctx, proposal)
		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutRequestDecryptionKeyPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutRequestDecryptionKeyPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data kstypes.RequestDecryptionKeyPacketData,
) error {
	pID, _ := strconv.ParseUint(data.GetProposalId(), 10, 64)
	proposal, found := k.GetProposal(ctx, pID)
	if !found {
		return errors.New("Proposal not found")
	}

	if (proposal.Status == v1.ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD) ||
		(proposal.Status == v1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD) {
		timeoutTimestamp := ctx.BlockTime().Add(time.Second * 20).UnixNano()

		_, _ = k.TransmitRequestDecryptionKeyPacket(ctx,
			data,
			packet.SourcePort,
			packet.SourceChannel,
			clienttypes.ZeroHeight(),
			uint64(timeoutTimestamp),
		)
	}

	return nil
}
