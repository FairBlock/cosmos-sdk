package v1

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govcodec "github.com/cosmos/cosmos-sdk/x/gov/codec"
	groupcodec "github.com/cosmos/cosmos-sdk/x/group/codec"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// governance module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgSubmitProposal{}, "cosmos-sdk/v1/MsgSubmitProposal")
	legacy.RegisterAminoMsg(cdc, &MsgDeposit{}, "cosmos-sdk/v1/MsgDeposit")
	legacy.RegisterAminoMsg(cdc, &MsgVote{}, "cosmos-sdk/v1/MsgVote")
	legacy.RegisterAminoMsg(cdc, &MsgVoteWeighted{}, "cosmos-sdk/v1/MsgVoteWeighted")
	legacy.RegisterAminoMsg(cdc, &MsgExecLegacyContent{}, "cosmos-sdk/v1/MsgExecLegacyContent")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/gov/v1/MsgUpdateParams")
}

// RegisterInterfaces registers the interfaces types with the Interface Registry.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSubmitProposal{},
		&MsgVote{},
		&MsgVoteWeighted{},
		&MsgDeposit{},
		&MsgExecLegacyContent{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

func init() {
	// Register all Amino interfaces and concrete types on the authz and gov Amino codec so that this can later be
	// used to properly serialize MsgGrant, MsgExec and MsgSubmitProposal instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
	RegisterLegacyAminoCodec(govcodec.Amino)
	RegisterLegacyAminoCodec(groupcodec.Amino)
}

// mustProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
// NOTE: Copied from https://github.com/cosmos/cosmos-sdk/blob/971c542453e0972ef1dfc5a80159ad5049c7211c/codec/json.go
// and modified in order to allow `EmitDefaults` to be set to false for ics20 packet marshalling.
// This allows for the introduction of the memo field to be backwards compatible.
func mustProtoMarshalJSON(msg proto.Message) []byte {
	anyResolver := codectypes.NewInterfaceRegistry()

	// EmitDefaults is set to false to prevent marshalling of unpopulated fields (memo)
	// OrigName and the anyResovler match the fields the original SDK function would expect
	// in order to minimize changes.

	// OrigName is true since there is no particular reason to use camel case
	// The any resolver is empty, but provided anyways.
	jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: false, AnyResolver: anyResolver}

	err := codectypes.UnpackInterfaces(msg, codectypes.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	if err := jm.Marshal(buf, msg); err != nil {
		panic(err)
	}

	return buf.Bytes()
}
