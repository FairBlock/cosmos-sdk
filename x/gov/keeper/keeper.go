package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"fmt"
	db "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/runtime"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"time"

	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	ibcKeeperFn        func() *ibckeeper.Keeper
	capabilityScopedFn func(string) capabilitykeeper.ScopedKeeper
	scopedKeeper       exported.ScopedKeeper

	authKeeper  types.AccountKeeper
	bankKeeper  types.BankKeeper
	distrKeeper types.DistributionKeeper

	// The reference to the DelegationSet and ValidatorSet to get information about validators and delegators
	sk types.StakingKeeper

	// GovHooks
	hooks types.GovHooks

	// The (unexposed) keys used to access the stores from the Context.
	storeService corestoretypes.KVStoreService

	// The codec for binary encoding/decoding.
	cdc codec.Codec

	// Legacy Proposal router
	legacyRouter v1beta1.Router

	// Msg server router
	router baseapp.MessageRouter

	config types.Config

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	Schema                 collections.Schema
	Constitution           collections.Item[string]
	Params                 collections.Item[v1.Params]
	Deposits               collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Deposit]
	Votes                  collections.Map[collections.Pair[uint64, sdk.AccAddress], v1.Vote]
	ProposalID             collections.Sequence
	Proposals              collections.Map[uint64, v1.Proposal]
	ActiveProposalsQueue   collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	InactiveProposalsQueue collections.Map[collections.Pair[time.Time, uint64], uint64] // TODO(tip): this should be simplified and go into an index.
	VotingPeriodProposals  collections.Map[uint64, []byte]                              // TODO(tip): this could be a keyset or index.
}

// GetAuthority returns the x/gov module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.Codec, storeService corestoretypes.KVStoreService, authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper, sk types.StakingKeeper, distrKeeper types.DistributionKeeper,
	router baseapp.MessageRouter, config types.Config, authority string,
	ibcKeeperFn func() *ibckeeper.Keeper,
	capabilityScopedFn func(string) capabilitykeeper.ScopedKeeper,
) *Keeper {
	// ensure governance module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	if _, err := authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	// If MaxMetadataLen not set by app developer, set to default value.
	if config.MaxMetadataLen == 0 {
		config.MaxMetadataLen = types.DefaultConfig().MaxMetadataLen
	}

	sb := collections.NewSchemaBuilder(storeService)
	k := &Keeper{
		storeService:           storeService,
		authKeeper:             authKeeper,
		bankKeeper:             bankKeeper,
		distrKeeper:            distrKeeper,
		sk:                     sk,
		cdc:                    cdc,
		router:                 router,
		config:                 config,
		authority:              authority,
		ibcKeeperFn:            ibcKeeperFn,
		capabilityScopedFn:     capabilityScopedFn,
		Constitution:           collections.NewItem(sb, types.ConstitutionKey, "constitution", collections.StringValue),
		Params:                 collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[v1.Params](cdc)),
		Deposits:               collections.NewMap(sb, types.DepositsKeyPrefix, "deposits", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Deposit](cdc)), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		Votes:                  collections.NewMap(sb, types.VotesKeyPrefix, "votes", collections.PairKeyCodec(collections.Uint64Key, sdk.LengthPrefixedAddressKey(sdk.AccAddressKey)), codec.CollValue[v1.Vote](cdc)),          // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
		ProposalID:             collections.NewSequence(sb, types.ProposalIDKey, "proposal_id"),
		Proposals:              collections.NewMap(sb, types.ProposalsKeyPrefix, "proposals", collections.Uint64Key, codec.CollValue[v1.Proposal](cdc)),
		ActiveProposalsQueue:   collections.NewMap(sb, types.ActiveProposalQueuePrefix, "active_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value),     // sdk.TimeKey is needed to retain state compatibility
		InactiveProposalsQueue: collections.NewMap(sb, types.InactiveProposalQueuePrefix, "inactive_proposals_queue", collections.PairKeyCodec(sdk.TimeKey, collections.Uint64Key), collections.Uint64Value), // sdk.TimeKey is needed to retain state compatibility
		VotingPeriodProposals:  collections.NewMap(sb, types.VotingPeriodProposalKeyPrefix, "voting_period_proposals", collections.Uint64Key, collections.BytesValue),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Hooks gets the hooks for governance *Keeper {
func (k *Keeper) Hooks() types.GovHooks {
	if k.hooks == nil {
		// return a no-op implementation if no hooks are set
		return types.MultiGovHooks{}
	}

	return k.hooks
}

// SetHooks sets the hooks for governance
func (k *Keeper) SetHooks(gh types.GovHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set governance hooks twice")
	}

	k.hooks = gh

	return k
}

// SetLegacyRouter sets the legacy router for governance
func (k *Keeper) SetLegacyRouter(router v1beta1.Router) {
	// It is vital to seal the governance proposal router here as to not allow
	// further handlers to be registered after the keeper is created since this
	// could create invalid or non-deterministic behavior.
	router.Seal()
	k.legacyRouter = router
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Router returns the gov keeper's router
func (k Keeper) Router() baseapp.MessageRouter {
	return k.router
}

// LegacyRouter returns the gov keeper's legacy router
func (k Keeper) LegacyRouter() v1beta1.Router {
	return k.legacyRouter
}

// GetGovernanceAccount returns the governance ModuleAccount
func (k Keeper) GetGovernanceAccount(ctx context.Context) sdk.ModuleAccountI {
	return k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
}

// ModuleAccountAddress returns gov module account address
func (k Keeper) ModuleAccountAddress() sdk.AccAddress {
	return k.authKeeper.GetModuleAddress(types.ModuleName)
}

// assertMetadataLength returns an error if given metadata length
// is greater than a pre-defined MaxMetadataLen.
func (k Keeper) assertMetadataLength(metadata string) error {
	if metadata != "" && uint64(len(metadata)) > k.config.MaxMetadataLen {
		return types.ErrMetadataTooLong.Wrapf("got metadata with length %d", len(metadata))
	}
	return nil
}

// assertSummaryLength returns an error if given summary length
// is greater than a pre-defined 40*MaxMetadataLen.
func (k Keeper) assertSummaryLength(summary string) error {
	if summary != "" && uint64(len(summary)) > 40*k.config.MaxMetadataLen {
		return types.ErrSummaryTooLong.Wrapf("got summary with length %d", len(summary))
	}
	return nil
}

// ProposalQueues

// InsertActiveProposalQueue inserts a proposalID into the active proposal queue at endTime
func (keeper Keeper) InsertActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.ActiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromActiveProposalQueue removes a proposalID from the Active Proposal Queue
func (keeper Keeper) RemoveFromActiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {

	store := keeper.storeService.OpenKVStore(ctx)
	store.Delete(types.ActiveProposalQueueKey(proposalID, endTime))
}

// InsertInactiveProposalQueue inserts a proposalID into the inactive proposal queue at endTime
func (keeper Keeper) InsertInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz := types.GetProposalIDBytes(proposalID)
	store.Set(types.InactiveProposalQueueKey(proposalID, endTime), bz)
}

// RemoveFromInactiveProposalQueue removes a proposalID from the Inactive Proposal Queue
func (keeper Keeper) RemoveFromInactiveProposalQueue(ctx sdk.Context, proposalID uint64, endTime time.Time) {
	store := keeper.storeService.OpenKVStore(ctx)
	store.Delete(types.InactiveProposalQueueKey(proposalID, endTime))
}

// Iterators

// IterateActiveProposalsQueue iterates over the proposals in the active proposal queue
// and performs a callback function
func (keeper Keeper) IterateActiveProposalsQueue(ctx sdk.Context, endTime time.Time, cb func(proposal v1.Proposal) (stop bool)) {
	iterator, err := keeper.ActiveProposalQueueIterator(ctx, endTime)
	if err != nil {
		fmt.Sprintf("error on iterating active proposal queue: %s", err.Error())
		return
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitActiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// IterateInactiveProposalsQueue iterates over the proposals in the inactive proposal queue
// and performs a callback function
func (keeper Keeper) IterateInactiveProposalsQueue(ctx sdk.Context, endTime time.Time, cb func(proposal v1.Proposal) (stop bool)) {
	iterator, err := keeper.InactiveProposalQueueIterator(ctx, endTime)
	if err != nil {
		fmt.Sprintf("error on iterating inactive proposal queue: %s", err.Error())
		return
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		proposalID, _ := types.SplitInactiveProposalQueueKey(iterator.Key())
		proposal, found := keeper.GetProposal(ctx, proposalID)
		if !found {
			panic(fmt.Sprintf("proposal %d does not exist", proposalID))
		}

		if cb(proposal) {
			break
		}
	}
}

// ActiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Active Queue that expire by endTime
func (keeper Keeper) ActiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) (db.Iterator, error) {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Iterator(types.ActiveProposalQueuePrefix, storetypes.PrefixEndBytes(types.ActiveProposalByTimeKey(endTime)))
}

// InactiveProposalQueueIterator returns an sdk.Iterator for all the proposals in the Inactive Queue that expire by endTime
func (keeper Keeper) InactiveProposalQueueIterator(ctx sdk.Context, endTime time.Time) (db.Iterator, error) {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Iterator(types.InactiveProposalQueuePrefix, storetypes.PrefixEndBytes(types.InactiveProposalByTimeKey(endTime)))
}

// ScopedKeeper returns the ScopedKeeper
func (k *Keeper) ScopedKeeper() exported.ScopedKeeper {
	if k.scopedKeeper == nil && k.capabilityScopedFn != nil {
		k.scopedKeeper = k.capabilityScopedFn(types.ModuleName)
	}
	return k.scopedKeeper
}

// ----------------------------------------------------------------------------
// IBC Keeper Logic
// ----------------------------------------------------------------------------

// ChanCloseInit defines a wrapper function for the channel Keeper's function.
func (k *Keeper) ChanCloseInit(ctx sdk.Context, portID, channelID string) error {
	capName := host.ChannelCapabilityPath(portID, channelID)
	chanCap, ok := k.ScopedKeeper().GetCapability(ctx, capName)
	if !ok {
		return errorsmod.Wrapf(channeltypes.ErrChannelCapabilityNotFound, "could not retrieve channel capability at: %s", capName)
	}
	return k.ibcKeeperFn().ChannelKeeper.ChanCloseInit(ctx, portID, channelID, chanCap)
}

// ShouldBound checks if the IBC app module can be bound to the desired port
func (k *Keeper) ShouldBound(ctx sdk.Context, portID string) bool {
	scopedKeeper := k.ScopedKeeper()
	if scopedKeeper == nil {
		return false
	}
	_, ok := scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return !ok
}

// BindPort defines a wrapper function for the port Keeper's function in
// order to expose it to module's InitGenesis function
func (k *Keeper) BindPort(ctx sdk.Context, portID string) error {
	cap := k.ibcKeeperFn().PortKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the IBC app module. Used in ExportGenesis
func (k *Keeper) GetPort(ctx sdk.Context) string {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	return string(store.Get(types.PortKey))
}

// SetPort sets the portID for the IBC app module. Used in InitGenesis
func (k *Keeper) SetPort(ctx sdk.Context, portID string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, []byte{})
	store.Set(types.PortKey, []byte(portID))
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k *Keeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return k.ScopedKeeper().AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the IBC app module to claim a capability that core IBC
// passes to it
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.ScopedKeeper().ClaimCapability(ctx, cap, name)
}
