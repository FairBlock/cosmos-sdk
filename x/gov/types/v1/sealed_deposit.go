package v1

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewSealedDeposit creates a new SealedDeposit instance
//
//nolint:interfacer
func NewSealedDeposit(proposalID uint64, depositor sdk.AccAddress, amount sdk.Coins) SealedDeposit {
	return SealedDeposit{proposalID, depositor.String(), amount}
}

// SealedDeposits is a collection of SealedDeposit objects
type SealedDeposits []*SealedDeposit

// Equal returns true if two slices (order-dependant) of deposits are equal.
func (d SealedDeposits) Equal(other SealedDeposits) bool {
	if len(d) != len(other) {
		return false
	}

	for i, deposit := range d {
		if deposit.String() != other[i].String() {
			return false
		}
	}

	return true
}

func (d SealedDeposits) String() string {
	if len(d) == 0 {
		return "[]"
	}
	out := fmt.Sprintf("Deposits for Proposal %d:", d[0].ProposalId)
	for _, dep := range d {
		out += fmt.Sprintf("\n  %s: %s", dep.Depositor, dep.Amount)
	}
	return out
}
