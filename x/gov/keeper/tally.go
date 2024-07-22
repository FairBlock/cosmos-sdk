package keeper

import (
	"bytes"
	"context"
	"encoding/hex"

	enc "github.com/FairBlock/DistributedIBE/encryption"
	bls "github.com/drand/kyber-bls12381"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TODO: Break into several smaller functions for clarity

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx context.Context, proposal v1.Proposal) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	results := make(map[v1.VoteOption]math.LegacyDec)
	results[v1.OptionYes] = math.LegacyZeroDec()
	results[v1.OptionAbstain] = math.LegacyZeroDec()
	results[v1.OptionNo] = math.LegacyZeroDec()
	results[v1.OptionNoWithVeto] = math.LegacyZeroDec()
	results[v1.OptionEncrypted] = math.LegacyZeroDec()

	totalVotingPower := math.LegacyZeroDec()
	currValidators := make(map[string]v1.ValidatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	err = keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		valBz, err := keeper.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			return false
		}
		currValidators[validator.GetOperator()] = v1.NewValidatorGovInfo(
			valBz,
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			math.LegacyZeroDec(),
			v1.WeightedVoteOptions{},
		)

		return false
	})
	if err != nil {
		return false, false, tallyResults, err
	}

	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
	err = keeper.Votes.Walk(ctx, rng, func(key collections.Pair[uint64, sdk.AccAddress], vote v1.Vote) (bool, error) {
		// if validator, just record it in the map
		voter, err := keeper.authKeeper.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			return false, err
		}

		valAddrStr, err := keeper.sk.ValidatorAddressCodec().BytesToString(voter)
		if err != nil {
			return false, err
		}
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Options
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		err = keeper.sk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will be deduction of voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
				currValidators[valAddrStr] = val

				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)

				for _, option := range vote.Options {
					weight, _ := math.LegacyNewDecFromStr(option.Weight)
					subPower := votingPower.Mul(weight)
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})
		if err != nil {
			return false, err
		}

		return false, keeper.Votes.Remove(ctx, collections.Join(vote.ProposalId, sdk.AccAddress(voter)))
	})

	if err != nil {
		return false, false, tallyResults, err
	}

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := sharesAfterDeductions.MulInt(val.BondedTokens).Quo(val.DelegatorShares)

		for _, option := range val.Vote {
			weight, _ := math.LegacyNewDecFromStr(option.Weight)
			subPower := votingPower.Mul(weight)
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return false, false, tallyResults, err
	}
	tallyResults = v1.NewTallyResultFromMap(results)

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	totalBonded, err := keeper.sk.TotalBondedTokens(ctx)
	if err != nil {
		return false, false, tallyResults, err
	}

	if totalBonded.IsZero() {
		return false, false, tallyResults, nil
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := math.LegacyNewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults, nil
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return false, false, tallyResults, nil
	}

	// If more than 1/3 of voters veto, proposal fails
	vetoThreshold, _ := math.LegacyNewDecFromStr(params.VetoThreshold)
	if results[v1.OptionNoWithVeto].Quo(totalVotingPower).GT(vetoThreshold) {
		return false, params.BurnVoteVeto, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	// For expedited 2/3
	var thresholdStr string
	if proposal.Expedited {
		thresholdStr = params.GetExpeditedThreshold()
	} else {
		thresholdStr = params.GetThreshold()
	}

	threshold, _ := math.LegacyNewDecFromStr(thresholdStr)

	if results[v1.OptionYes].Quo(totalVotingPower.Sub(results[v1.OptionAbstain])).GT(threshold) {
		return true, false, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults, nil
}

// DecryptVotes decrypts any encrypted votes
func (keeper Keeper) DecryptVotes(ctx sdk.Context, proposal v1.Proposal) {
	pubKey := proposal.Pubkey
	publicKeyByte, _ := hex.DecodeString(pubKey)
	suite := bls.NewBLS12381Suite()

	publicKeyPoint := suite.G1().Point()
	publicKeyPoint.UnmarshalBinary(publicKeyByte)

	keyByte, _ := hex.DecodeString(proposal.AggrKeyshare)

	skPoint := suite.G2().Point()
	skPoint.UnmarshalBinary(keyByte)

	var deletedVotes, modifiedVotes []v1.Vote

	keeper.IterateVotes(ctx, proposal.Id, func(vote v1.Vote) bool {
		if vote.Options[0].Option == v1.OptionEncrypted {
			if vote.EncryptedVoteData != "" {
				var decryptedVote bytes.Buffer
				var voteBuffer bytes.Buffer

				decodedVoteData, err := hex.DecodeString(vote.EncryptedVoteData)
				if err != nil {
					deletedVotes = append(deletedVotes, vote)
					return false
				}

				_, err = voteBuffer.Write(decodedVoteData)

				if err != nil {
					deletedVotes = append(deletedVotes, vote)
					return false
				}

				err = enc.Decrypt(publicKeyPoint, skPoint, &decryptedVote, &voteBuffer)
				if err != nil {
					deletedVotes = append(deletedVotes, vote)
					return false
				}

				var decVote v1.DecryptedVoteOption
				err = decVote.Unmarshal(decryptedVote.Bytes())
				if err != nil {
					deletedVotes = append(deletedVotes, vote)
					return false
				}

				if decVote.Option == v1.OptionEncrypted {
					deletedVotes = append(deletedVotes, vote)
					return false
				}

				vote.Options[0].Option = decVote.Option
				vote.Options[0].Weight = "1"

				modifiedVotes = append(modifiedVotes, vote)

				return false
			}
			deletedVotes = append(deletedVotes, vote)
			return false
		}
		return false
	})

	for _, dv := range deletedVotes {
		voter := sdk.MustAccAddressFromBech32(dv.Voter)
		keeper.Votes.Remove(ctx, collections.Join(dv.ProposalId, voter))
	}

	for _, mv := range modifiedVotes {
		voter := sdk.MustAccAddressFromBech32(mv.Voter)
		keeper.Votes.Set(ctx, collections.Join(mv.ProposalId, voter), mv)
	}
}
