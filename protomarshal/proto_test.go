package protomarshal_test

import (
	"bytes"
	"testing"

	"github.com/gopencloud/grpc-api-gateway/internal/testpb"
	"github.com/gopencloud/grpc-api-gateway/protomarshal"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var message = &testpb.ABitOfEverything{
	SingleNested:        &testpb.ABitOfEverything_Nested{},
	RepeatedStringValue: nil,
	MappedStringValue:   nil,
	MappedNestedValue:   nil,
	RepeatedEnumValue:   nil,
	TimestampValue:      &timestamppb.Timestamp{},
	Uuid:                "6EC2446F-7E89-4127-B3E6-5C05E6BECBA7",
	Nested: []*testpb.ABitOfEverything_Nested{
		{
			Name:   "foo",
			Amount: 12345,
		},
	},
	Uint64Value: 0xFFFFFFFFFFFFFFFF,
	EnumValue:   testpb.NumericEnum_ONE,
	OneofValue: &testpb.ABitOfEverything_OneofString{
		OneofString: "bar",
	},
	MapValue: map[string]testpb.NumericEnum{
		"a": testpb.NumericEnum_ONE,
		"b": testpb.NumericEnum_ZERO,
	},
}

func TestProtoMarshalUnmarshal(t *testing.T) {
	marshaller := protomarshal.ProtoMarshaller{}

	// Marshal
	buffer, err := marshaller.Marshal(message)
	if err != nil {
		t.Fatalf("Marshalling returned error: %s", err.Error())
	}

	// Unmarshal
	unmarshalled := &testpb.ABitOfEverything{}
	err = marshaller.Unmarshal(buffer, unmarshalled)
	if err != nil {
		t.Fatalf("Unmarshalling returned error: %s", err.Error())
	}

	if !proto.Equal(unmarshalled, message) {
		t.Errorf(
			"Unmarshalled didn't match original message: (original = %v) != (unmarshalled = %v)",
			unmarshalled,
			message,
		)
	}
}

func TestProtoEncoderDecodert(t *testing.T) {
	marshaller := protomarshal.ProtoMarshaller{}

	var buf bytes.Buffer

	encoder := marshaller.NewEncoder(&buf)
	decoder := marshaller.NewDecoder(&buf)

	// Encode
	err := encoder.Encode(message)
	if err != nil {
		t.Fatalf("Encoding returned error: %s", err.Error())
	}

	// Decode
	unencoded := &testpb.ABitOfEverything{}
	err = decoder.Decode(unencoded)
	if err != nil {
		t.Fatalf("Unmarshalling returned error: %s", err.Error())
	}

	if !proto.Equal(unencoded, message) {
		t.Errorf(
			"Unencoded didn't match original message: (original = %v) != (unencoded = %v)",
			unencoded,
			message,
		)
	}
}
