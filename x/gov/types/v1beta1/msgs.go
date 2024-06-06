package v1beta1

import (
	"fmt"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/codec"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"
	"sigs.k8s.io/yaml"
)

// Governance message types and routes
const (
	TypeMsgDeposit        = "deposit"
	TypeMsgVote           = "vote"
	TypeMsgVoteEncrypted  = "encrypted_vote"
	TypeMsgVoteWeighted   = "weighted_vote"
	TypeMsgSubmitProposal = "submit_proposal"
)

var (
	_, _, _, _ sdk.Msg = &MsgSubmitProposal{}, &MsgDeposit{}, &MsgVote{}, &MsgVoteWeighted{}

	_ codectypes.UnpackInterfacesMessage = &MsgSubmitProposal{}
)

// NewMsgSubmitProposal creates a new MsgSubmitProposal.
func NewMsgSubmitProposal(content Content, initialDeposit sdk.Coins, proposer sdk.AccAddress) (*MsgSubmitProposal, error) {
	m := &MsgSubmitProposal{
		InitialDeposit: initialDeposit,
		Proposer:       proposer.String(),
	}
	err := m.SetContent(content)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetInitialDeposit returns the initial deposit of MsgSubmitProposal.
func (m *MsgSubmitProposal) GetInitialDeposit() sdk.Coins { return m.InitialDeposit }

// GetContent returns the content of MsgSubmitProposal.
func (m *MsgSubmitProposal) GetContent() Content {
	content, ok := m.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

// SetInitialDeposit sets the given initial deposit for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetInitialDeposit(coins sdk.Coins) {
	m.InitialDeposit = coins
}

// SetProposer sets the given proposer address for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetProposer(address fmt.Stringer) {
	m.Proposer = address.String()
}

// SetContent sets the content for MsgSubmitProposal.
func (m *MsgSubmitProposal) SetContent(content Content) error {
	msg, ok := content.(proto.Message)
	if !ok {
		return fmt.Errorf("can't proto marshal %T", msg)
	}
	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}
	m.Content = any
	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m MsgSubmitProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(m.Content, &content)
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(depositor sdk.AccAddress, proposalID uint64, amount sdk.Coins) *MsgDeposit {
	return &MsgDeposit{proposalID, depositor.String(), amount}
}

// NewMsgVote creates a message to cast a vote on an active proposal
func NewMsgVote(voter sdk.AccAddress, proposalID uint64, option VoteOption) *MsgVote {
	return &MsgVote{proposalID, voter.String(), option}
}

// NewMsgVoteEncrypted creates a message to cast an encrypted vote on an active proposal
//
//nolint:interfacer
func NewMsgVoteEncrypted(voter sdk.AccAddress, proposalID uint64, encData string) *MsgVoteEncrypted {
	return &MsgVoteEncrypted{proposalID, voter.String(), encData}
}

// Route implements the sdk.Msg interface.
func (msg MsgVoteEncrypted) Route() string { return types.RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgVoteEncrypted) Type() string { return TypeMsgVoteEncrypted }

// ValidateBasic implements the sdk.Msg interface.
func (msg MsgVoteEncrypted) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Voter); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid voter address: %s", err)
	}

	return nil
}

// String implements the Stringer interface
func (msg MsgVoteEncrypted) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgVoteEncrypted) GetSignBytes() []byte {
	bz := codec.ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the expected signers for a MsgVote.
func (msg MsgVoteEncrypted) GetSigners() []sdk.AccAddress {
	voter, _ := sdk.AccAddressFromBech32(msg.Voter)
	return []sdk.AccAddress{voter}
}

// NewMsgVoteWeighted creates a message to cast a vote on an active proposal.
func NewMsgVoteWeighted(voter sdk.AccAddress, proposalID uint64, options WeightedVoteOptions) *MsgVoteWeighted {
	return &MsgVoteWeighted{proposalID, voter.String(), options}
}
