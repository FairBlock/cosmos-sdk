package keeper

import (
	"context"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/collections"
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AddVote adds a vote on a specific proposal
func (keeper Keeper) AddVote(ctx context.Context, proposalID uint64, voterAddr sdk.AccAddress, options v1.WeightedVoteOptions, encData string, metadata string) error {
	// Check if proposal is in voting period.
	inVotingPeriod, err := keeper.VotingPeriodProposals.Has(ctx, proposalID)
	if err != nil {
		return err
	}

	if !inVotingPeriod {
		return errors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	err = keeper.assertMetadataLength(metadata)
	if err != nil {
		return err
	}

	for _, option := range options {
		if !v1.ValidWeightedVoteOption(*option) {
			return errors.Wrap(types.ErrInvalidVote, option.String())
		}
	}

	vote := v1.NewVote(proposalID, voterAddr, options, encData, metadata)
	err = keeper.Votes.Set(ctx, collections.Join(proposalID, voterAddr), vote)
	if err != nil {
		return err
	}

	if encData != "" {
		proposal, _ := keeper.GetProposal(ctx, proposalID)
		if !proposal.HasEncryptedVotes {
			proposal.HasEncryptedVotes = true
			keeper.SetProposal(ctx, proposal)
		}
	}

	// called after a vote on a proposal is cast
	err = keeper.Hooks().AfterProposalVote(ctx, proposalID, voterAddr)
	if err != nil {
		return err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalVote,
			sdk.NewAttribute(types.AttributeKeyVoter, voterAddr.String()),
			sdk.NewAttribute(types.AttributeKeyOption, options.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	return nil
}

// SetVote sets a Vote to the gov store
func (keeper Keeper) SetVote(ctx sdk.Context, vote v1.Vote) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz := keeper.cdc.MustMarshal(&vote)
	addr := sdk.MustAccAddressFromBech32(vote.Voter)

	store.Set(types.VoteKey(vote.ProposalId, addr), bz)
}

// IterateVotes iterates over all the proposals votes and performs a callback function
func (keeper Keeper) IterateVotes(ctx sdk.Context, proposalID uint64, cb func(vote v1.Vote) (stop bool)) {
	store := keeper.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.VotesKey(proposalID))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var vote v1.Vote
		keeper.cdc.MustUnmarshal(iterator.Value(), &vote)

		if cb(vote) {
			break
		}
	}
}

// deleteVotes deletes all the votes from a given proposalID.
func (keeper Keeper) deleteVotes(ctx context.Context, proposalID uint64) error {
	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposalID)
	err := keeper.Votes.Clear(ctx, rng)
	if err != nil {
		return err
	}

	return nil
}

// deleteVote deletes a vote from a given proposalID and voter from the store
func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) {
	store := keeper.storeService.OpenKVStore(ctx)
	store.Delete(types.VoteKey(proposalID, voterAddr))
}
