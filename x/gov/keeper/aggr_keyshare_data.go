package keeper

import (
	"errors"
	"fmt"
	"strconv"

	"fairyring/x/keyshare/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// OnRecvAggrKeyshareDataPacket processes packet reception
func (k Keeper) OnRecvAggrKeyshareDataPacket(ctx sdk.Context, packet channeltypes.Packet, data types.AggrKeyshareDataPacketData) (packetAck types.AggrKeyshareDataPacketAck, err error) {
	fmt.Println("\n\n\nOnRecvAggrKeyshareDataPacket\n\n\n")

	// validate packet data upon receiving
	if err := data.ValidateBasic(); err != nil {
		return packetAck, err
	}

	fmt.Println("\n\n\n\n")
	fmt.Println(data.AggrKeyshare)
	fmt.Println(data.Identity)
	fmt.Println(data.ProposalId)

	pID, err := strconv.ParseUint(data.ProposalId, 10, 64)
	if err != nil {
		fmt.Println("PARSE ERROR")
	}

	fmt.Println("PID: ", pID)

	proposal, found := k.GetProposal(ctx, pID)
	if !found {
		fmt.Println("proposal not found")
		return packetAck, errors.New("Proposal not found")
	}

	fmt.Println("\n\n\n\n")

	proposal.AggrKeyshare = data.AggrKeyshare
	k.SetProposal(ctx, proposal)

	return packetAck, nil
}

// OnAcknowledgementAggrKeyshareDataPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementAggrKeyshareDataPacket(ctx sdk.Context, packet channeltypes.Packet, data types.AggrKeyshareDataPacketData, ack channeltypes.Acknowledgement) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:

		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error

		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.AggrKeyshareDataPacketAck

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

// OnTimeoutAggrKeyshareDataPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutAggrKeyshareDataPacket(ctx sdk.Context, packet channeltypes.Packet, data types.AggrKeyshareDataPacketData) error {

	// TODO: packet timeout logic

	return nil
}
