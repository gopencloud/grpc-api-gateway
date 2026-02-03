package protopath_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gopencloud/grpc-api-gateway/internal/testpb"
	"github.com/gopencloud/grpc-api-gateway/protopath"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestPopulateFromPath(t *testing.T) {
	testCases := []struct {
		Name    string
		Message proto.Message
		Path    string
		Value   string
		Result  proto.Message
		Error   bool
	}{
		{
			Name:    "Enum",
			Message: &testpb.Proto3Message{},
			Path:    "enum_value",
			Value:   "Y",
			Result:  &testpb.Proto3Message{EnumValue: testpb.EnumValue_Y},
		},
		{
			Name:    "Nested",
			Message: &testpb.Proto3Message{},
			Path:    "nested.double_value",
			Value:   "5.6",
			Result:  &testpb.Proto3Message{Nested: &testpb.Proto3Message{DoubleValue: 5.6}},
		},
		{
			Name:    "NestedDeep",
			Message: &testpb.Proto3Message{},
			Path:    "nested.nested.repeated_enum",
			Value:   "Z",
			Result: &testpb.Proto3Message{
				Nested: &testpb.Proto3Message{
					Nested: &testpb.Proto3Message{
						RepeatedEnum: []testpb.EnumValue{testpb.EnumValue_Z},
					},
				},
			},
		},
		{
			Name:    "OneOf",
			Message: &testpb.Proto3Message{},
			Path:    "oneof_bool_value",
			Value:   "true",
			Result: &testpb.Proto3Message{
				OneofValue: &testpb.Proto3Message_OneofBoolValue{OneofBoolValue: true},
			},
		},
		{
			Name: "OneOf",
			Message: &testpb.Proto3Message{
				OneofValue: &testpb.Proto3Message_OneofBoolValue{OneofBoolValue: true},
			},
			Path:  "oneof_bool_value",
			Value: "true",
			Result: &testpb.Proto3Message{
				OneofValue: &testpb.Proto3Message_OneofStringValue{OneofStringValue: "string_value"},
			},
			Error: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			if err := protopath.PopulateFieldFromPath(tt.Message, tt.Path, tt.Value); err != nil {
				if !tt.Error {
					t.Fatalf("received error when expected no error: %v", err)
				}
				return
			}
			if tt.Error {
				t.Fatal("expected error but received none")
				return
			}

			if diff := cmp.Diff(tt.Message, tt.Result, protocmp.Transform()); diff != "" {
				t.Fatalf("unexpected result: %s", diff)
			}
		})
	}
}
