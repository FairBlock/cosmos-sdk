package keeper

import (
	"errors"
	"fmt"

	"github.com/Fairblock/fairyring/x/keyshare/types"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

// TransmitGetDecryptionKeyPacket transmits the packet over IBC with the specified source port and source channel
func (k Keeper) TransmitGetDecryptionKeyPacket(
	ctx sdk.Context,
	packetData types.GetDecryptionKeyPacketData,
	sourcePort,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
) (uint64, error) {
	channelCap, ok := k.ScopedKeeper().GetCapability(
		ctx,
		host.ChannelCapabilityPath(sourcePort, sourceChannel),
	)
	if !ok {
		return 0, sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	packetBytes := packetData.GetBytes()

	return k.ibcKeeperFn().ChannelKeeper.SendPacket(ctx, channelCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, packetBytes)
}

// OnAcknowledgementGetDecryptionKeyPacket responds to the the success or failure of a packet
// acknowledgement written on the receiving chain.
func (k Keeper) OnAcknowledgementGetDecryptionKeyPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.GetDecryptionKeyPacketData,
	ack channeltypes.Acknowledgement,
) error {
	switch dispatchedAck := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		fmt.Println("\n\n\n ACK Err: ", dispatchedAck.Error, "\n\n\n")
		// TODO: failed acknowledgement logic
		_ = dispatchedAck.Error
		return nil
	case *channeltypes.Acknowledgement_Result:
		// Decode the packet acknowledgment
		var packetAck types.GetDecryptionKeyPacketAck

		if err := types.ModuleCdc.UnmarshalJSON(dispatchedAck.Result, &packetAck); err != nil {
			// The counter-party module doesn't implement the correct acknowledgment format
			return errors.New("cannot unmarshal acknowledgment")
		}

		return nil
	default:
		// The counter-party module doesn't implement the correct acknowledgment format
		return errors.New("invalid acknowledgment format")
	}
}

// OnTimeoutGetDecryptionKeyPacket responds to the case where a packet has not been transmitted because of a timeout
func (k Keeper) OnTimeoutGetDecryptionKeyPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data types.GetDecryptionKeyPacketData,
) error {

	// No processing is required since GetAggrKeysharePacket is sent
	// every block till a response is received or the tally period is over.
	// ref: abci.go: 108

	return nil
}
