// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.28.0--rc2
// source: proto/control.proto

package protocol

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type MouseEvent_EventType int32

const (
	MouseEvent_MOVE        MouseEvent_EventType = 0
	MouseEvent_LEFT_DOWN   MouseEvent_EventType = 1
	MouseEvent_LEFT_UP     MouseEvent_EventType = 2
	MouseEvent_RIGHT_DOWN  MouseEvent_EventType = 3
	MouseEvent_RIGHT_UP    MouseEvent_EventType = 4
	MouseEvent_MIDDLE_DOWN MouseEvent_EventType = 5
	MouseEvent_MIDDLE_UP   MouseEvent_EventType = 6
	MouseEvent_SCROLL      MouseEvent_EventType = 7
)

// Enum value maps for MouseEvent_EventType.
var (
	MouseEvent_EventType_name = map[int32]string{
		0: "MOVE",
		1: "LEFT_DOWN",
		2: "LEFT_UP",
		3: "RIGHT_DOWN",
		4: "RIGHT_UP",
		5: "MIDDLE_DOWN",
		6: "MIDDLE_UP",
		7: "SCROLL",
	}
	MouseEvent_EventType_value = map[string]int32{
		"MOVE":        0,
		"LEFT_DOWN":   1,
		"LEFT_UP":     2,
		"RIGHT_DOWN":  3,
		"RIGHT_UP":    4,
		"MIDDLE_DOWN": 5,
		"MIDDLE_UP":   6,
		"SCROLL":      7,
	}
)

func (x MouseEvent_EventType) Enum() *MouseEvent_EventType {
	p := new(MouseEvent_EventType)
	*p = x
	return p
}

func (x MouseEvent_EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MouseEvent_EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_control_proto_enumTypes[0].Descriptor()
}

func (MouseEvent_EventType) Type() protoreflect.EnumType {
	return &file_proto_control_proto_enumTypes[0]
}

func (x MouseEvent_EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MouseEvent_EventType.Descriptor instead.
func (MouseEvent_EventType) EnumDescriptor() ([]byte, []int) {
	return file_proto_control_proto_rawDescGZIP(), []int{0, 0}
}

type KeyEvent_EventType int32

const (
	KeyEvent_KEY_DOWN KeyEvent_EventType = 0
	KeyEvent_KEY_UP   KeyEvent_EventType = 1
)

// Enum value maps for KeyEvent_EventType.
var (
	KeyEvent_EventType_name = map[int32]string{
		0: "KEY_DOWN",
		1: "KEY_UP",
	}
	KeyEvent_EventType_value = map[string]int32{
		"KEY_DOWN": 0,
		"KEY_UP":   1,
	}
)

func (x KeyEvent_EventType) Enum() *KeyEvent_EventType {
	p := new(KeyEvent_EventType)
	*p = x
	return p
}

func (x KeyEvent_EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (KeyEvent_EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_control_proto_enumTypes[1].Descriptor()
}

func (KeyEvent_EventType) Type() protoreflect.EnumType {
	return &file_proto_control_proto_enumTypes[1]
}

func (x KeyEvent_EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use KeyEvent_EventType.Descriptor instead.
func (KeyEvent_EventType) EnumDescriptor() ([]byte, []int) {
	return file_proto_control_proto_rawDescGZIP(), []int{1, 0}
}

type MouseEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventType   MouseEvent_EventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=godesk.MouseEvent_EventType" json:"event_type,omitempty"`
	X           int32                `protobuf:"varint,2,opt,name=x,proto3" json:"x,omitempty"`
	Y           int32                `protobuf:"varint,3,opt,name=y,proto3" json:"y,omitempty"`
	ScrollDelta int32                `protobuf:"varint,4,opt,name=scroll_delta,json=scrollDelta,proto3" json:"scroll_delta,omitempty"` // 用于滚轮事件
}

func (x *MouseEvent) Reset() {
	*x = MouseEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_control_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MouseEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MouseEvent) ProtoMessage() {}

func (x *MouseEvent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_control_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MouseEvent.ProtoReflect.Descriptor instead.
func (*MouseEvent) Descriptor() ([]byte, []int) {
	return file_proto_control_proto_rawDescGZIP(), []int{0}
}

func (x *MouseEvent) GetEventType() MouseEvent_EventType {
	if x != nil {
		return x.EventType
	}
	return MouseEvent_MOVE
}

func (x *MouseEvent) GetX() int32 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *MouseEvent) GetY() int32 {
	if x != nil {
		return x.Y
	}
	return 0
}

func (x *MouseEvent) GetScrollDelta() int32 {
	if x != nil {
		return x.ScrollDelta
	}
	return 0
}

type KeyEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EventType KeyEvent_EventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=godesk.KeyEvent_EventType" json:"event_type,omitempty"`
	KeyCode   int32              `protobuf:"varint,2,opt,name=key_code,json=keyCode,proto3" json:"key_code,omitempty"`
	Shift     bool               `protobuf:"varint,3,opt,name=shift,proto3" json:"shift,omitempty"`
	Ctrl      bool               `protobuf:"varint,4,opt,name=ctrl,proto3" json:"ctrl,omitempty"`
	Alt       bool               `protobuf:"varint,5,opt,name=alt,proto3" json:"alt,omitempty"`
	Meta      bool               `protobuf:"varint,6,opt,name=meta,proto3" json:"meta,omitempty"` // Windows键或Command键
}

func (x *KeyEvent) Reset() {
	*x = KeyEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_control_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KeyEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KeyEvent) ProtoMessage() {}

func (x *KeyEvent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_control_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KeyEvent.ProtoReflect.Descriptor instead.
func (*KeyEvent) Descriptor() ([]byte, []int) {
	return file_proto_control_proto_rawDescGZIP(), []int{1}
}

func (x *KeyEvent) GetEventType() KeyEvent_EventType {
	if x != nil {
		return x.EventType
	}
	return KeyEvent_KEY_DOWN
}

func (x *KeyEvent) GetKeyCode() int32 {
	if x != nil {
		return x.KeyCode
	}
	return 0
}

func (x *KeyEvent) GetShift() bool {
	if x != nil {
		return x.Shift
	}
	return false
}

func (x *KeyEvent) GetCtrl() bool {
	if x != nil {
		return x.Ctrl
	}
	return false
}

func (x *KeyEvent) GetAlt() bool {
	if x != nil {
		return x.Alt
	}
	return false
}

func (x *KeyEvent) GetMeta() bool {
	if x != nil {
		return x.Meta
	}
	return false
}

var File_proto_control_proto protoreflect.FileDescriptor

var file_proto_control_proto_rawDesc = []byte{
	0x0a, 0x13, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x67, 0x6f, 0x64, 0x65, 0x73, 0x6b, 0x22, 0x85, 0x02,
	0x0a, 0x0a, 0x4d, 0x6f, 0x75, 0x73, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x3b, 0x0a, 0x0a,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1c, 0x2e, 0x67, 0x6f, 0x64, 0x65, 0x73, 0x6b, 0x2e, 0x4d, 0x6f, 0x75, 0x73, 0x65, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x01, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x63, 0x72, 0x6f, 0x6c, 0x6c, 0x5f,
	0x64, 0x65, 0x6c, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73, 0x63, 0x72,
	0x6f, 0x6c, 0x6c, 0x44, 0x65, 0x6c, 0x74, 0x61, 0x22, 0x7b, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x4d, 0x4f, 0x56, 0x45, 0x10, 0x00, 0x12,
	0x0d, 0x0a, 0x09, 0x4c, 0x45, 0x46, 0x54, 0x5f, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x12, 0x0b,
	0x0a, 0x07, 0x4c, 0x45, 0x46, 0x54, 0x5f, 0x55, 0x50, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x52,
	0x49, 0x47, 0x48, 0x54, 0x5f, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x52,
	0x49, 0x47, 0x48, 0x54, 0x5f, 0x55, 0x50, 0x10, 0x04, 0x12, 0x0f, 0x0a, 0x0b, 0x4d, 0x49, 0x44,
	0x44, 0x4c, 0x45, 0x5f, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x05, 0x12, 0x0d, 0x0a, 0x09, 0x4d, 0x49,
	0x44, 0x44, 0x4c, 0x45, 0x5f, 0x55, 0x50, 0x10, 0x06, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x43, 0x52,
	0x4f, 0x4c, 0x4c, 0x10, 0x07, 0x22, 0xd7, 0x01, 0x0a, 0x08, 0x4b, 0x65, 0x79, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x64, 0x65, 0x73, 0x6b, 0x2e,
	0x4b, 0x65, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x19, 0x0a,
	0x08, 0x6b, 0x65, 0x79, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x07, 0x6b, 0x65, 0x79, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x68, 0x69, 0x66,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x73, 0x68, 0x69, 0x66, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x74, 0x72, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x63, 0x74,
	0x72, 0x6c, 0x12, 0x10, 0x0a, 0x03, 0x61, 0x6c, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x03, 0x61, 0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x04, 0x6d, 0x65, 0x74, 0x61, 0x22, 0x25, 0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0c, 0x0a, 0x08, 0x4b, 0x45, 0x59, 0x5f, 0x44, 0x4f, 0x57,
	0x4e, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x4b, 0x45, 0x59, 0x5f, 0x55, 0x50, 0x10, 0x01, 0x42,
	0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x75,
	0x69, 0x66, 0x65, 0x69, 0x2f, 0x67, 0x6f, 0x64, 0x65, 0x73, 0x6b, 0x2f, 0x69, 0x6e, 0x74, 0x65,
	0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_control_proto_rawDescOnce sync.Once
	file_proto_control_proto_rawDescData = file_proto_control_proto_rawDesc
)

func file_proto_control_proto_rawDescGZIP() []byte {
	file_proto_control_proto_rawDescOnce.Do(func() {
		file_proto_control_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_control_proto_rawDescData)
	})
	return file_proto_control_proto_rawDescData
}

var file_proto_control_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_control_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_control_proto_goTypes = []any{
	(MouseEvent_EventType)(0), // 0: godesk.MouseEvent.EventType
	(KeyEvent_EventType)(0),   // 1: godesk.KeyEvent.EventType
	(*MouseEvent)(nil),        // 2: godesk.MouseEvent
	(*KeyEvent)(nil),          // 3: godesk.KeyEvent
}
var file_proto_control_proto_depIdxs = []int32{
	0, // 0: godesk.MouseEvent.event_type:type_name -> godesk.MouseEvent.EventType
	1, // 1: godesk.KeyEvent.event_type:type_name -> godesk.KeyEvent.EventType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_proto_control_proto_init() }
func file_proto_control_proto_init() {
	if File_proto_control_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_control_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*MouseEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_proto_control_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*KeyEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_control_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_control_proto_goTypes,
		DependencyIndexes: file_proto_control_proto_depIdxs,
		EnumInfos:         file_proto_control_proto_enumTypes,
		MessageInfos:      file_proto_control_proto_msgTypes,
	}.Build()
	File_proto_control_proto = out.File
	file_proto_control_proto_rawDesc = nil
	file_proto_control_proto_goTypes = nil
	file_proto_control_proto_depIdxs = nil
}
