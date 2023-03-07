package signing

import (
	"fmt"
	"strings"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// GetSignersContext is a context for retrieving the list of signers from a
// message where signers are specified by the cosmos.msg.v1.signer protobuf
// option.
type GetSignersContext struct {
	protoFiles      *protoregistry.Files
	protoTypes      protoregistry.MessageTypeResolver
	getSignersFuncs map[protoreflect.FullName]getSignersFunc
}

// GetSignersOptions are options for creating GetSignersContext.
type GetSignersOptions struct {
	// ProtoFiles are the protobuf files to use for resolving message descriptors.
	// If it is nil, the global protobuf registry will be used.
	ProtoFiles *protoregistry.Files

	ProtoTypes protoregistry.MessageTypeResolver
}

// NewGetSignersContext creates a new GetSignersContext using the provided options.
func NewGetSignersContext(options GetSignersOptions) *GetSignersContext {
	protoFiles := options.ProtoFiles
	if protoFiles == nil {
		protoFiles = protoregistry.GlobalFiles
	}

	protoTypes := options.ProtoTypes
	if protoTypes == nil {
		protoTypes = protoregistry.GlobalTypes
	}

	return &GetSignersContext{
		protoFiles:      protoFiles,
		protoTypes:      protoTypes,
		getSignersFuncs: map[protoreflect.FullName]getSignersFunc{},
	}
}

type getSignersFunc func(proto.Message) []string

func getSignersFieldNames(descriptor protoreflect.MessageDescriptor) ([]string, error) {
	signersFields := proto.GetExtension(descriptor.Options(), msgv1.E_Signer).([]string)
	if signersFields == nil || len(signersFields) == 0 {
		return nil, fmt.Errorf("no cosmos.msg.v1.signer option found for message %s", descriptor.FullName())
	}

	return signersFields, nil
}

func (*GetSignersContext) makeGetSignersFunc(descriptor protoreflect.MessageDescriptor) (getSignersFunc, error) {
	signersFields, err := getSignersFieldNames(descriptor)
	if err != nil {
		return nil, err
	}

	fieldGetters := make([]func(proto.Message, []string) []string, len(signersFields))
	for i, fieldName := range signersFields {
		field := descriptor.Fields().ByName(protoreflect.Name(fieldName))
		if field == nil {
			return nil, fmt.Errorf("field %s not found in message %s", fieldName, descriptor.FullName())
		}

		if field.IsMap() || field.HasOptionalKeyword() {
			return nil, fmt.Errorf("cosmos.msg.v1.signer field %s in message %s must not be a map or optional", fieldName, descriptor.FullName())
		}

		switch field.Kind() {
		case protoreflect.StringKind:
			if field.IsList() {
				fieldGetters[i] = func(msg proto.Message, arr []string) []string {
					signers := msg.ProtoReflect().Get(field).List()
					n := signers.Len()
					for i := 0; i < n; i++ {
						arr = append(arr, signers.Get(i).String())
					}
					return arr
				}
			} else {
				fieldGetters[i] = func(msg proto.Message, arr []string) []string {
					return append(arr, msg.ProtoReflect().Get(field).String())
				}
			}
		case protoreflect.MessageKind:
			isList := field.IsList()
			nestedMessage := field.Message()
			nestedSignersFields, err := getSignersFieldNames(nestedMessage)
			if err != nil {
				return nil, err
			}

			if len(nestedSignersFields) != 1 {
				return nil, fmt.Errorf("nested cosmos.msg.v1.signer option in message %s must contain only one value", nestedMessage.FullName())
			}

			nestedFieldName := nestedSignersFields[0]
			nestedField := nestedMessage.Fields().ByName(protoreflect.Name(nestedFieldName))
			nestedIsList := nestedField.IsList()
			if nestedField == nil {
				return nil, fmt.Errorf("field %s not found in message %s", nestedFieldName, nestedMessage.FullName())
			}

			if nestedField.Kind() != protoreflect.StringKind || nestedField.IsMap() || nestedField.HasOptionalKeyword() {
				return nil, fmt.Errorf("nested signer field %s in message %s must be a simple string", nestedFieldName, nestedMessage.FullName())
			}

			if isList {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							signers := msgs.Get(i).Message().Get(nestedField).List()
							n := signers.Len()
							for j := 0; j < n; j++ {
								arr = append(arr, signers.Get(j).String())
							}
						}
						return arr
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						msgs := msg.ProtoReflect().Get(field).List()
						m := msgs.Len()
						for i := 0; i < m; i++ {
							arr = append(arr, msgs.Get(i).Message().Get(nestedField).String())
						}
						return arr
					}
				}
			} else {
				if nestedIsList {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						nestedMsg := msg.ProtoReflect().Get(field).Message()
						signers := nestedMsg.Get(nestedField).List()
						n := signers.Len()
						for j := 0; j < n; j++ {
							arr = append(arr, signers.Get(j).String())
						}
						return arr
					}
				} else {
					fieldGetters[i] = func(msg proto.Message, arr []string) []string {
						return append(arr, msg.ProtoReflect().Get(field).Message().Get(nestedField).String())
					}
				}
			}

		default:
			return nil, fmt.Errorf("unexpected field type %s for field %s in message %s", field.Kind(), fieldName, descriptor.FullName())
		}
	}

	return func(message proto.Message) []string {
		var signers []string
		for _, getter := range fieldGetters {
			signers = getter(message, signers)
		}
		return signers
	}, nil
}

// GetSigners returns the signers for a given message.
func (c *GetSignersContext) GetSigners(msg proto.Message) ([]string, error) {
	messageDescriptor := msg.ProtoReflect().Descriptor()
	f, ok := c.getSignersFuncs[messageDescriptor.FullName()]
	if !ok {
		var err error
		f, err = c.makeGetSignersFunc(messageDescriptor)
		if err != nil {
			return nil, err
		}
		c.getSignersFuncs[messageDescriptor.FullName()] = f
	}

	return f(msg), nil
}

// GetSignersForAny returns the signers for a given proto message wrapped inside
// an Any. It can be used to retrieve the signers for a message which does not
// have a concrete protoreflect type in the current binary because it uses dynamicpb
// under the hood. This is useful for usage in dynamic clients (like autocli) or
// for working with legacy gogo protobuf messages.
func (c *GetSignersContext) GetSignersForAny(any *anypb.Any) ([]string, error) {
	url := any.TypeUrl
	typ, err := c.protoTypes.FindMessageByURL(url)
	if err != nil {
		if err == protoregistry.NotFound {
			msgName := protoreflect.FullName(url)
			if i := strings.LastIndexByte(url, '/'); i >= 0 {
				msgName = msgName[i+len("/"):]
			}
			msgDesc, err := c.protoFiles.FindDescriptorByName(msgName)
			if err != nil {
				return nil, err
			}

			typ = dynamicpb.NewMessageType(msgDesc.(protoreflect.MessageDescriptor))
		} else {
			return nil, err
		}
	}

	msg := typ.New().Interface()
	err = any.UnmarshalTo(msg)
	if err != nil {
		return nil, err
	}

	return c.GetSigners(msg)
}
