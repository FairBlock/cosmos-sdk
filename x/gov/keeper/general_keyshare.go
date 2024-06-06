package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	commontypes "github.com/Fairblock/fairyring/x/common/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetRequestQueueEntry returns a queue entry by its identity
func (k Keeper) GetRequestQueueEntry(
	ctx sdk.Context,
	proposalID string,
) (val commontypes.RequestAggrKeyshare, found bool) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.ReqQueueKeyPrefix))

	b := store.Get(types.QueueKey(
		proposalID,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// SetQueueEntry sets a queue entry by its identity
func (k Keeper) SetReqQueueEntry(
	ctx sdk.Context,
	val commontypes.RequestAggrKeyshare,
) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.ReqQueueKeyPrefix))

	entry := k.cdc.MustMarshal(&val)
	store.Set(
		types.QueueKey(val.GetProposalId()),
		entry,
	)
}

// RemoveQueueEntry removes an entry from the store
func (k Keeper) RemoveReqQueueEntry(
	ctx sdk.Context,
	proposalID string,
) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.ReqQueueKeyPrefix))
	store.Delete(types.QueueKey(proposalID))
}

// GetAllGenEncTxQueueEntry returns all GenEncTxQueue entries
func (k Keeper) GetAllReqQueueEntry(ctx sdk.Context) (list []commontypes.RequestAggrKeyshare) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.ReqQueueKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.RequestAggrKeyshare
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}

// GetQueueEntry returns a queue entry by its identity
func (k Keeper) GetSignalQueueEntry(
	ctx sdk.Context,
	proposalID string,
) (val commontypes.GetAggrKeyshare, found bool) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.SignalQueueKeyPrefix))
	b := store.Get(types.QueueKey(
		proposalID,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// SetQueueEntry sets a queue entry by its identity
func (k Keeper) SetSignalQueueEntry(
	ctx sdk.Context,
	val commontypes.GetAggrKeyshare,
) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.SignalQueueKeyPrefix))
	entry := k.cdc.MustMarshal(&val)
	store.Set(
		types.QueueKey(val.GetProposalId()),
		entry,
	)
}

// RemoveQueueEntry removes an entry from the store
func (k Keeper) RemoveSignalQueueEntry(
	ctx sdk.Context,
	proposalID string,
) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.SignalQueueKeyPrefix))
	store.Delete(types.QueueKey(proposalID))
}

// GetAllGenEncTxQueueEntry returns all GenEncTxQueue entries
func (k Keeper) GetAllSignalQueueEntry(ctx sdk.Context) (list []commontypes.GetAggrKeyshare) {
	store := prefix.NewStore(runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx)), types.KeyPrefix(types.SignalQueueKeyPrefix))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val commontypes.GetAggrKeyshare
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}
	return
}
