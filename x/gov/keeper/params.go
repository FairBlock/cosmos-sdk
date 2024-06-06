package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SetParams sets the gov module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params v1.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetParams gets the gov module's parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params v1.Params) {
	store := k.storeService.OpenKVStore(clientCtx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil || bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}
