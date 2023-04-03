package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// GetSealedDeposit gets the deposit of a specific depositor on a specific sealed proposal
func (keeper Keeper) GetSealedDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) (deposit v1.SealedDeposit, found bool) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get(types.SealedDepositKey(proposalID, depositorAddr))
	if bz == nil {
		return deposit, false
	}

	keeper.cdc.MustUnmarshal(bz, &deposit)

	return deposit, true
}

// SetSealedDeposit sets a Deposit to the gov store
func (keeper Keeper) SetSealedDeposit(ctx sdk.Context, deposit v1.SealedDeposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshal(&deposit)
	depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

	store.Set(types.SealedDepositKey(deposit.ProposalId, depositor), bz)
}

// GetAllSealedDeposits returns all the deposits from the store
func (keeper Keeper) GetAllSealedDeposits(ctx sdk.Context) (deposits v1.SealedDeposits) {
	keeper.IterateAllSealedDeposits(ctx, func(deposit v1.SealedDeposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})

	return
}

// GetSealedDeposits returns all the sealed deposits from a proposal
func (keeper Keeper) GetSealedDeposits(ctx sdk.Context, proposalID uint64) (deposits v1.SealedDeposits) {
	keeper.IterateSealedDeposits(ctx, proposalID, func(deposit v1.SealedDeposit) bool {
		deposits = append(deposits, &deposit)
		return false
	})

	return
}

// DeleteAndBurnSealedDeposits deletes and burn all the deposits on a specific proposal.
func (keeper Keeper) DeleteAndBurnSealedDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateSealedDeposits(ctx, proposalID, func(deposit v1.SealedDeposit) bool {
		err := keeper.bankKeeper.BurnCoins(ctx, types.ModuleName, deposit.Amount)
		if err != nil {
			panic(err)
		}

		depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

		store.Delete(types.SealedDepositKey(proposalID, depositor))
		return false
	})
}

// IterateAllSealedDeposits iterates over all the stored deposits and performs a callback function
func (keeper Keeper) IterateAllSealedDeposits(ctx sdk.Context, cb func(deposit v1.SealedDeposit) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.SealedDepositsKeyPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.SealedDeposit

		keeper.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// IterateSealedDeposits iterates over all the proposals deposits and performs a callback function
func (keeper Keeper) IterateSealedDeposits(ctx sdk.Context, proposalID uint64, cb func(deposit v1.SealedDeposit) (stop bool)) {
	store := ctx.KVStore(keeper.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.SealedDepositsKey(proposalID))

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var deposit v1.Deposit

		keeper.cdc.MustUnmarshal(iterator.Value(), &deposit)

		if cb(deposit) {
			break
		}
	}
}

// AddSealedDeposit adds or updates a deposit of a specific depositor on a specific proposal.
// Activates voting period when appropriate and returns true in that case, else returns false.
func (keeper Keeper) AddSealedDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress, depositAmount sdk.Coins) (bool, error) {
	// Checks to see if proposal exists
	proposal, ok := keeper.GetSealedProposal(ctx, proposalID)
	if !ok {
		return false, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}

	// Check if proposal is still depositable
	if (proposal.Status != v1.StatusDepositPeriod) && (proposal.Status != v1.StatusVotingPeriod) {
		return false, sdkerrors.Wrapf(types.ErrInactiveProposal, "%d", proposalID)
	}

	// update the governance module's account coins pool
	err := keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, depositorAddr, types.ModuleName, depositAmount)
	if err != nil {
		return false, err
	}

	// Update proposal
	proposal.TotalDeposit = sdk.NewCoins(proposal.TotalDeposit...).Add(depositAmount...)
	keeper.SetSealedProposal(ctx, proposal)

	// Check if deposit has provided sufficient total funds to transition the proposal into the voting period
	activatedVotingPeriod := false

	if proposal.Status == v1.StatusDepositPeriod && sdk.NewCoins(proposal.TotalDeposit...).IsAllGTE(keeper.GetDepositParams(ctx).MinDeposit) {
		keeper.ActivateSecretVotingPeriod(ctx, proposal)

		activatedVotingPeriod = true
	}

	// Add or update deposit object
	deposit, found := keeper.GetSealedDeposit(ctx, proposalID, depositorAddr)

	if found {
		deposit.Amount = sdk.NewCoins(deposit.Amount...).Add(depositAmount...)
	} else {
		deposit = v1.NewSealedDeposit(proposalID, depositorAddr, depositAmount)
	}

	// called when deposit has been added to a proposal, however the proposal may not be active
	keeper.AfterProposalDeposit(ctx, proposalID, depositorAddr)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeProposalDeposit,
			sdk.NewAttribute(sdk.AttributeKeyAmount, depositAmount.String()),
			sdk.NewAttribute(types.AttributeKeyProposalID, fmt.Sprintf("%d", proposalID)),
		),
	)

	keeper.SetSealedDeposit(ctx, deposit)

	return activatedVotingPeriod, nil
}

// RefundAndDeleteSealedDeposits refunds and deletes all the deposits on a specific proposal.
func (keeper Keeper) RefundAndDeleteSealedDeposits(ctx sdk.Context, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)

	keeper.IterateSealedDeposits(ctx, proposalID, func(deposit v1.SealedDeposit) bool {
		depositor := sdk.MustAccAddressFromBech32(deposit.Depositor)

		err := keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, depositor, deposit.Amount)
		if err != nil {
			panic(err)
		}

		store.Delete(types.SealedDepositKey(proposalID, depositor))
		return false
	})
}
