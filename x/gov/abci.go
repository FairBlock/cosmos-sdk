package gov

import (
	"encoding/hex"
	"fmt"
	"time"

	kstypes "fairyring/x/keyshare/types"

	distIBE "github.com/FairBlock/DistributedIBE"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	bls "github.com/drand/kyber-bls12381"
)

// EndBlocker called every block, process inflation, update validator set.
func EndBlocker(ctx sdk.Context, keeper *keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	logger := keeper.Logger(ctx)

	// delete dead proposals from store and returns theirs deposits.
	// A proposal is dead when it's inactive and didn't get enough deposit on time to get into voting phase.
	keeper.IterateInactiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal v1.Proposal) bool {
		keeper.DeleteProposal(ctx, proposal.Id)

		params := keeper.GetParams(ctx)
		if !params.BurnProposalDepositPrevote {
			keeper.RefundAndDeleteDeposits(ctx, proposal.Id) // refund deposit if proposal got removed without getting 100% of the proposal
		} else {
			keeper.DeleteAndBurnDeposits(ctx, proposal.Id) // burn the deposit if proposal got removed without getting 100% of the proposal
		}

		// called when proposal become inactive
		keeper.Hooks().AfterProposalFailedMinDeposit(ctx, proposal.Id)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeInactiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, types.AttributeValueProposalDropped),
			),
		)

		logger.Info(
			"proposal did not meet minimum deposit; deleted",
			"proposal", proposal.Id,
			"min_deposit", sdk.NewCoins(params.MinDeposit...).String(),
			"total_deposit", sdk.NewCoins(proposal.TotalDeposit...).String(),
		)

		return false
	})

	// fetch active proposals whose voting periods have ended (are passed the block time)
	keeper.IterateActiveProposalsQueue(ctx, ctx.BlockHeader().Time, func(proposal v1.Proposal) bool {
		var tagValue, logMsg string
		tallyPeriod := keeper.GetParams(ctx).MaxTallyPeriod
		tallyEndTime := proposal.VotingEndTime.Add(*tallyPeriod)

		if proposal.HasEncryptedVotes {
			if proposal.AggrKeyshare == "" && (ctx.BlockTime().Compare(tallyEndTime) >= 0) {
				proposal.Status = v1.StatusFailed
				tagValue = types.AttributeValueProposalFailed
				logMsg = "failed"

				keeper.RefundAndDeleteDeposits(ctx, proposal.Id)

				keeper.SetProposal(ctx, proposal)
				keeper.RemoveFromActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)

				// when proposal become active
				keeper.Hooks().AfterProposalVotingPeriodEnded(ctx, proposal.Id)

				logger.Info(
					"proposal tallied",
					"proposal", proposal.Id,
					"results", logMsg,
				)

				ctx.EventManager().EmitEvent(
					sdk.NewEvent(
						types.EventTypeActiveProposal,
						sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
						sdk.NewAttribute(types.AttributeKeyProposalResult, tagValue),
					),
				)
				return false
			}

			if proposal.Status == v1.StatusVotingPeriod {
				proposal.Status = v1.StatusTallyPeriod
				keeper.SetProposal(ctx, proposal)
			}

			if proposal.AggrKeyshare == "" {
				var packetData kstypes.GetAggrKeysharePacketData
				sPort := keeper.GetPort(ctx)
				params := keeper.GetParams(ctx)
				packetData.Identity = proposal.Identity
				timeoutTimestamp := ctx.BlockTime().Add(time.Second * 20).UnixNano()

				_, err := keeper.TransmitGetAggrKeysharePacket(ctx,
					packetData,
					sPort,
					params.ChannelId,
					clienttypes.ZeroHeight(),
					uint64(timeoutTimestamp),
				)

				if err != nil {
					logger.Info(
						"IBC Request to fetch aggr. Keyshare failed",
						"proposal", proposal.Id,
						"error", err,
					)
				}

				//===========================================//
				// FOR TESTING ONLY, HARDCODE AGGR. KEYSHARE //
				//===========================================//

				fmt.Println("\n\n\nHardcoding aggr keyshare\n\n\n")
				shareByte, err := hex.DecodeString("29c861be5016b20f5a4397795e3f086d818b11ad02e0dd8ee28e485988b6cb07")
				if err != nil {
					fmt.Println("invalid share provided")
					return false
				}

				share := bls.NewKyberScalar()
				err = share.UnmarshalBinary(shareByte)
				if err != nil {
					fmt.Printf("invalid share provided, got error while unmarshaling: %s\n", err.Error())
					return false
				}

				s := bls.NewBLS12381Suite()
				extractedKey := distIBE.Extract(s, share, 1, []byte(proposal.Identity))

				extractedBinary, err := extractedKey.SK.MarshalBinary()
				if err != nil {
					fmt.Printf("Error while marshaling the extracted key: %s\n", err.Error())
					return false
				}
				extractedKeyHex := hex.EncodeToString(extractedBinary)
				proposal.AggrKeyshare = extractedKeyHex
				keeper.SetProposal(ctx, proposal)
				fmt.Println("\n\n\nProposal Set\n\n\n")

				return false
			}

			keeper.DecryptVotes(ctx, proposal)
		}

		passes, burnDeposits, tallyResults := keeper.Tally(ctx, proposal)
		fmt.Println("\n\n\nvotes tallied\n\n\n")

		if burnDeposits {
			keeper.DeleteAndBurnDeposits(ctx, proposal.Id)
		} else {
			keeper.RefundAndDeleteDeposits(ctx, proposal.Id)
		}

		if passes {
			var (
				idx    int
				events sdk.Events
				msg    sdk.Msg
			)

			// attempt to execute all messages within the passed proposal
			// Messages may mutate state thus we use a cached context. If one of
			// the handlers fails, no state mutation is written and the error
			// message is logged.
			cacheCtx, writeCache := ctx.CacheContext()
			messages, err := proposal.GetMsgs()
			if err == nil {
				for idx, msg = range messages {
					handler := keeper.Router().Handler(msg)

					var res *sdk.Result
					res, err = handler(cacheCtx, msg)
					if err != nil {
						break
					}

					events = append(events, res.GetEvents()...)
				}
			}

			// `err == nil` when all handlers passed.
			// Or else, `idx` and `err` are populated with the msg index and error.
			if err == nil {
				proposal.Status = v1.StatusPassed
				tagValue = types.AttributeValueProposalPassed
				logMsg = "passed"

				// write state to the underlying multi-store
				writeCache()

				// propagate the msg events to the current context
				ctx.EventManager().EmitEvents(events)
			} else {
				proposal.Status = v1.StatusFailed
				tagValue = types.AttributeValueProposalFailed
				logMsg = fmt.Sprintf("passed, but msg %d (%s) failed on execution: %s", idx, sdk.MsgTypeURL(msg), err)
			}
		} else {
			proposal.Status = v1.StatusRejected
			tagValue = types.AttributeValueProposalRejected
			logMsg = "rejected"
		}

		proposal.FinalTallyResult = &tallyResults

		keeper.SetProposal(ctx, proposal)
		keeper.RemoveFromActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)

		// when proposal become active
		keeper.Hooks().AfterProposalVotingPeriodEnded(ctx, proposal.Id)

		logger.Info(
			"proposal tallied",
			"proposal", proposal.Id,
			"results", logMsg,
		)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeActiveProposal,
				sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposal.Id)),
				sdk.NewAttribute(types.AttributeKeyProposalResult, tagValue),
			),
		)
		return false
	})
}
