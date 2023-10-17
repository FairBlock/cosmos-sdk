package keeper

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/math"
	enc "github.com/FairBlock/DistributedIBE/encryption"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	bls "github.com/drand/kyber-bls12381"
)

// TODO: Break into several smaller functions for clarity

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx sdk.Context, proposal v1.Proposal) (passes bool, burnDeposits bool, tallyResults v1.TallyResult) {
	fmt.Println("\n\n\nTallying votes\n\n\n")
	results := make(map[v1.VoteOption]sdk.Dec)
	results[v1.OptionYes] = math.LegacyZeroDec()
	results[v1.OptionAbstain] = math.LegacyZeroDec()
	results[v1.OptionNo] = math.LegacyZeroDec()
	results[v1.OptionNoWithVeto] = math.LegacyZeroDec()

	totalVotingPower := math.LegacyZeroDec()
	currValidators := make(map[string]v1.ValidatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		currValidators[validator.GetOperator().String()] = v1.NewValidatorGovInfo(
			validator.GetOperator(),
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			math.LegacyZeroDec(),
			v1.WeightedVoteOptions{},
		)

		return false
	})

	keeper.IterateVotes(ctx, proposal.Id, func(vote v1.Vote) bool {
		// if validator, just record it in the map
		voter := sdk.MustAccAddressFromBech32(vote.Voter)

		valAddrStr := sdk.ValAddress(voter.Bytes()).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Options
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		keeper.sk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr().String()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will be deduction of voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
				currValidators[valAddrStr] = val

				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)

				for _, option := range vote.Options {
					weight, _ := sdk.NewDecFromStr(option.Weight)
					subPower := votingPower.Mul(weight)
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})

		keeper.deleteVote(ctx, vote.ProposalId, voter)
		return false
	})

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := sharesAfterDeductions.MulInt(val.BondedTokens).Quo(val.DelegatorShares)

		for _, option := range val.Vote {
			weight, _ := sdk.NewDecFromStr(option.Weight)
			subPower := votingPower.Mul(weight)
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	params := keeper.GetParams(ctx)
	tallyResults = v1.NewTallyResultFromMap(results)

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	if keeper.sk.TotalBondedTokens(ctx).IsZero() {
		return false, false, tallyResults
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(sdk.NewDecFromInt(keeper.sk.TotalBondedTokens(ctx)))
	quorum, _ := sdk.NewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return false, false, tallyResults
	}

	// If more than 1/3 of voters veto, proposal fails
	vetoThreshold, _ := sdk.NewDecFromStr(params.VetoThreshold)
	if results[v1.OptionNoWithVeto].Quo(totalVotingPower).GT(vetoThreshold) {
		return false, params.BurnVoteVeto, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	threshold, _ := sdk.NewDecFromStr(params.Threshold)
	if results[v1.OptionYes].Quo(totalVotingPower.Sub(results[v1.OptionAbstain])).GT(threshold) {
		return true, false, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults
}

// DecryptVotes decrypts any encrypted votes
func (keeper Keeper) DecryptVotes(ctx sdk.Context, proposal v1.Proposal) {
	fmt.Println("\n\n\nDecrypting Votes\n\n\n")

	pubKey := proposal.Pubkey
	publicKeyByte, _ := hex.DecodeString(pubKey)
	fmt.Println("1")

	suite := bls.NewBLS12381Suite()

	publicKeyPoint := suite.G1().Point()
	publicKeyPoint.UnmarshalBinary(publicKeyByte)

	keyByte, _ := hex.DecodeString(proposal.AggrKeyshare)
	fmt.Println("2")

	skPoint := suite.G2().Point()
	skPoint.UnmarshalBinary(keyByte)

	var deletedVotes, modifiedVotes []v1.Vote

	keeper.IterateVotes(ctx, proposal.Id, func(vote v1.Vote) bool {
		fmt.Println("3")
		fmt.Println("Options :", vote.Options)
		if vote.Options[0].Option == v1.OptionEncrypted {
			if vote.EncryptedVoteData != "" {
				fmt.Println("Vote : ", vote.EncryptedVoteData)
				var decryptedVote bytes.Buffer
				var voteBuffer bytes.Buffer
				_, err := voteBuffer.Write([]byte(vote.EncryptedVoteData))

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
				}

				vote.Options[0].Option = decVote.Option
				vote.Options[0].Weight = "1"

				modifiedVotes = append(modifiedVotes, vote)

				return false
			}
		}
		return false
	})

	fmt.Println("\n\n\nModified Votes:\n", modifiedVotes)
	fmt.Println("\n\n\nDeleted Votes:\n", deletedVotes)

	for _, dv := range deletedVotes {
		voter := sdk.MustAccAddressFromBech32(dv.Voter)
		keeper.deleteVote(ctx, dv.ProposalId, voter)
	}

	for _, mv := range modifiedVotes {
		keeper.SetVote(ctx, mv)
	}
}
