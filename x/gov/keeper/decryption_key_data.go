package keeper

import (
	"errors"
	"strconv"

	"github.com/Fairblock/fairyring/x/keyshare/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

// OnRecvDecryptionKeyDataPacket processes packet reception
func (k Keeper) OnRecvDecryptionKeyDataPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.DecryptionKeyDataPacketData,
) (packetAck types.DecryptionKeyPacketAck, err error) {
	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	pID, err := strconv.ParseUint(data.ProposalId, 10, 64)
	if err != nil {
		return packetAck, err
	}

	proposal, found := k.GetProposal(ctx, pID)
	if !found {
		return packetAck, errors.New("Proposal not found")
	}

	proposal.DecryptionKey = data.DecryptionKey
	err = k.SetProposal(ctx, proposal)
	if err != nil {
		return packetAck, err
	}

	return packetAck, nil
}

// OnAcknowledgementDecryptionKeyDataPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementDecryptionKeyDataPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.DecryptionKeyDataPacketData,
	ack channeltypes.Acknowledgement,
) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:

		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error

		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.DecryptionKeyPacketAck

		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		// TODO: successful acknowledgement logic

		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutDecryptionKeyDataPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutDecryptionKeyDataPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.DecryptionKeyDataPacketData,
) error {

	// TODO: packet timeout logic

	return nil
}

func (k Keeper) ProcessAggrKeyshare(ctx sdk.Context, pID string, aggrKeyshare string) error {
	id, err := strconv.ParseUint(pID, 10, 64)
	if err != nil {
		return err
	}

	proposal, found := k.GetProposal(ctx, id)
	if !found {
		return errors.New("Proposal not found")
	}

	proposal.DecryptionKey = aggrKeyshare
	k.SetProposal(ctx, proposal)
	return nil
}
