// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vdb.proto

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	vdb.proto

It has these top-level messages:
	Command
	Response
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Command_Type int32

const (
	Command_UNKNOWN Command_Type = 0
	Command_TIME    Command_Type = 1
	Command_GET     Command_Type = 2
	Command_SET     Command_Type = 3
	Command_REMOVE  Command_Type = 4
)

var Command_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "TIME",
	2: "GET",
	3: "SET",
	4: "REMOVE",
}
var Command_Type_value = map[string]int32{
	"UNKNOWN": 0,
	"TIME":    1,
	"GET":     2,
	"SET":     3,
	"REMOVE":  4,
}

func (x Command_Type) String() string {
	return proto.EnumName(Command_Type_name, int32(x))
}
func (Command_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Command struct {
	Type  Command_Type `protobuf:"varint,1,opt,name=type,enum=api.Command_Type" json:"type,omitempty"`
	ID    string       `protobuf:"bytes,2,opt,name=ID" json:"ID,omitempty"`
	Value string       `protobuf:"bytes,3,opt,name=value" json:"value,omitempty"`
}

func (m *Command) Reset()                    { *m = Command{} }
func (m *Command) String() string            { return proto.CompactTextString(m) }
func (*Command) ProtoMessage()               {}
func (*Command) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Command) GetType() Command_Type {
	if m != nil {
		return m.Type
	}
	return Command_UNKNOWN
}

func (m *Command) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Command) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type Response struct {
	ID    string `protobuf:"bytes,1,opt,name=ID" json:"ID,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value" json:"value,omitempty"`
	Error string `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Response) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *Response) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *Response) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*Command)(nil), "api.Command")
	proto.RegisterType((*Response)(nil), "api.Response")
	proto.RegisterEnum("api.Command_Type", Command_Type_name, Command_Type_value)
}

func init() { proto.RegisterFile("vdb.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 206 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2c, 0x4b, 0x49, 0xd2,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4e, 0x2c, 0xc8, 0x54, 0x9a, 0xcc, 0xc8, 0xc5, 0xee,
	0x9c, 0x9f, 0x9b, 0x9b, 0x98, 0x97, 0x22, 0xa4, 0xca, 0xc5, 0x52, 0x52, 0x59, 0x90, 0x2a, 0xc1,
	0xa8, 0xc0, 0xa8, 0xc1, 0x67, 0x24, 0xa8, 0x97, 0x58, 0x90, 0xa9, 0x07, 0x95, 0xd3, 0x0b, 0xa9,
	0x2c, 0x48, 0x0d, 0x02, 0x4b, 0x0b, 0xf1, 0x71, 0x31, 0x79, 0xba, 0x48, 0x30, 0x29, 0x30, 0x6a,
	0x70, 0x06, 0x31, 0x79, 0xba, 0x08, 0x89, 0x70, 0xb1, 0x96, 0x25, 0xe6, 0x94, 0xa6, 0x4a, 0x30,
	0x83, 0x85, 0x20, 0x1c, 0x25, 0x6b, 0x2e, 0x16, 0x90, 0x1e, 0x21, 0x6e, 0x2e, 0xf6, 0x50, 0x3f,
	0x6f, 0x3f, 0xff, 0x70, 0x3f, 0x01, 0x06, 0x21, 0x0e, 0x2e, 0x96, 0x10, 0x4f, 0x5f, 0x57, 0x01,
	0x46, 0x21, 0x76, 0x2e, 0x66, 0x77, 0xd7, 0x10, 0x01, 0x26, 0x10, 0x23, 0xd8, 0x35, 0x44, 0x80,
	0x59, 0x88, 0x8b, 0x8b, 0x2d, 0xc8, 0xd5, 0xd7, 0x3f, 0xcc, 0x55, 0x80, 0x45, 0xc9, 0x8d, 0x8b,
	0x23, 0x28, 0xb5, 0xb8, 0x20, 0x3f, 0xaf, 0x18, 0x66, 0x1d, 0x23, 0xa6, 0x75, 0x4c, 0x48, 0xd6,
	0x81, 0x44, 0x53, 0x8b, 0x8a, 0xf2, 0x8b, 0x60, 0x8e, 0x00, 0x73, 0x92, 0xd8, 0xc0, 0x3e, 0x35,
	0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x9d, 0x20, 0x45, 0x57, 0xf6, 0x00, 0x00, 0x00,
}
