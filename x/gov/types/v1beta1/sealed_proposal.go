package v1beta1

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"sigs.k8s.io/yaml"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewSealedProposal(content Content, id uint64, submitTime, depositEndTime time.Time) (SealedProposal, error) {
	msg, ok := content.(proto.Message)
	if !ok {
		return SealedProposal{}, fmt.Errorf("%T does not implement proto.Message", content)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return SealedProposal{}, err
	}

	p := SealedProposal{
		Content:          any,
		ProposalId:       id,
		Status:           StatusDepositPeriod,
		FinalTallyResult: EmptyTallyResult(),
		TotalDeposit:     sdk.NewCoins(),
		SubmitTime:       submitTime,
		DepositEndTime:   depositEndTime,
	}

	return p, nil
}

// String implements stringer interface
func (p SealedProposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetContent returns the proposal Content
func (p SealedProposal) GetContent() Content {
	content, ok := p.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

func (p SealedProposal) ProposalType() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalType()
}

func (p SealedProposal) ProposalRoute() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalRoute()
}

func (p SealedProposal) GetTitle() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.GetTitle()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p SealedProposal) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(p.Content, &content)
}

// SealedProposals is an array of Sealed proposal
type SealedProposals []SealedProposal

var _ codectypes.UnpackInterfacesMessage = SealedProposals{}

// Equal returns true if two slices (order-dependant) of proposals are equal.
func (p SealedProposals) Equal(other SealedProposals) bool {
	if len(p) != len(other) {
		return false
	}

	for i, proposal := range p {
		if !proposal.Equal(other[i]) {
			return false
		}
	}

	return true
}

// String implements stringer interface
func (p SealedProposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - (%s) [%s] %s\n",
			prop.ProposalId, prop.Status,
			prop.ProposalType(), prop.GetTitle())
	}
	return strings.TrimSpace(out)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p SealedProposals) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, x := range p {
		err := x.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}
