// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.1
// 	protoc        v5.29.2
// source: game.proto

package shared

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

// アイテムのステータス
type ItemStatus int32

const (
	ItemStatus_ACTIVE  ItemStatus = 0
	ItemStatus_REMOVED ItemStatus = 1
)

// Enum value maps for ItemStatus.
var (
	ItemStatus_name = map[int32]string{
		0: "ACTIVE",
		1: "REMOVED",
	}
	ItemStatus_value = map[string]int32{
		"ACTIVE":  0,
		"REMOVED": 1,
	}
)

func (x ItemStatus) Enum() *ItemStatus {
	p := new(ItemStatus)
	*p = x
	return p
}

func (x ItemStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ItemStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_game_proto_enumTypes[0].Descriptor()
}

func (ItemStatus) Type() protoreflect.EnumType {
	return &file_game_proto_enumTypes[0]
}

func (x ItemStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ItemStatus.Descriptor instead.
func (ItemStatus) EnumDescriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{0}
}

// 向き
type Direction int32

const (
	Direction_UP    Direction = 0
	Direction_DOWN  Direction = 1
	Direction_LEFT  Direction = 2
	Direction_RIGHT Direction = 3
)

// Enum value maps for Direction.
var (
	Direction_name = map[int32]string{
		0: "UP",
		1: "DOWN",
		2: "LEFT",
		3: "RIGHT",
	}
	Direction_value = map[string]int32{
		"UP":    0,
		"DOWN":  1,
		"LEFT":  2,
		"RIGHT": 3,
	}
)

func (x Direction) Enum() *Direction {
	p := new(Direction)
	*p = x
	return p
}

func (x Direction) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Direction) Descriptor() protoreflect.EnumDescriptor {
	return file_game_proto_enumTypes[1].Descriptor()
}

func (Direction) Type() protoreflect.EnumType {
	return &file_game_proto_enumTypes[1]
}

func (x Direction) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Direction.Descriptor instead.
func (Direction) EnumDescriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{1}
}

// プレイヤーのステータス
type Status int32

const (
	Status_ALIVE        Status = 0
	Status_DISCONNECTED Status = 1
)

// Enum value maps for Status.
var (
	Status_name = map[int32]string{
		0: "ALIVE",
		1: "DISCONNECTED",
	}
	Status_value = map[string]int32{
		"ALIVE":        0,
		"DISCONNECTED": 1,
	}
)

func (x Status) Enum() *Status {
	p := new(Status)
	*p = x
	return p
}

func (x Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Status) Descriptor() protoreflect.EnumDescriptor {
	return file_game_proto_enumTypes[2].Descriptor()
}

func (Status) Type() protoreflect.EnumType {
	return &file_game_proto_enumTypes[2]
}

func (x Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Status.Descriptor instead.
func (Status) EnumDescriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{2}
}

// アイテムの種類
type ItemType int32

const (
	ItemType_BULLET ItemType = 0
)

// Enum value maps for ItemType.
var (
	ItemType_name = map[int32]string{
		0: "BULLET",
	}
	ItemType_value = map[string]int32{
		"BULLET": 0,
	}
)

func (x ItemType) Enum() *ItemType {
	p := new(ItemType)
	*p = x
	return p
}

func (x ItemType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ItemType) Descriptor() protoreflect.EnumDescriptor {
	return file_game_proto_enumTypes[3].Descriptor()
}

func (ItemType) Type() protoreflect.EnumType {
	return &file_game_proto_enumTypes[3]
}

func (x ItemType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ItemType.Descriptor instead.
func (ItemType) EnumDescriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{3}
}

// 位置情報
type Position struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	X             int32                  `protobuf:"varint,1,opt,name=x,proto3" json:"x,omitempty"`
	Y             int32                  `protobuf:"varint,2,opt,name=y,proto3" json:"y,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Position) Reset() {
	*x = Position{}
	mi := &file_game_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Position) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Position) ProtoMessage() {}

func (x *Position) ProtoReflect() protoreflect.Message {
	mi := &file_game_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Position.ProtoReflect.Descriptor instead.
func (*Position) Descriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{0}
}

func (x *Position) GetX() int32 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *Position) GetY() int32 {
	if x != nil {
		return x.Y
	}
	return 0
}

// プレイヤーの状態
type PlayerState struct {
	state     protoimpl.MessageState `protogen:"open.v1"`
	PlayerId  string                 `protobuf:"bytes,1,opt,name=player_id,json=playerId,proto3" json:"player_id,omitempty"`
	Position  *Position              `protobuf:"bytes,2,opt,name=position,proto3" json:"position,omitempty"`
	Direction Direction              `protobuf:"varint,4,opt,name=direction,proto3,enum=terminalshooter.Direction" json:"direction,omitempty"`
	// statusはserverからのみ送信する
	Status        Status `protobuf:"varint,3,opt,name=status,proto3,enum=terminalshooter.Status" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PlayerState) Reset() {
	*x = PlayerState{}
	mi := &file_game_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlayerState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlayerState) ProtoMessage() {}

func (x *PlayerState) ProtoReflect() protoreflect.Message {
	mi := &file_game_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PlayerState.ProtoReflect.Descriptor instead.
func (*PlayerState) Descriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{1}
}

func (x *PlayerState) GetPlayerId() string {
	if x != nil {
		return x.PlayerId
	}
	return ""
}

func (x *PlayerState) GetPosition() *Position {
	if x != nil {
		return x.Position
	}
	return nil
}

func (x *PlayerState) GetDirection() Direction {
	if x != nil {
		return x.Direction
	}
	return Direction_UP
}

func (x *PlayerState) GetStatus() Status {
	if x != nil {
		return x.Status
	}
	return Status_ALIVE
}

// アイテムの状態
type ItemState struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	ItemId   string                 `protobuf:"bytes,1,opt,name=item_id,json=itemId,proto3" json:"item_id,omitempty"`
	Type     ItemType               `protobuf:"varint,2,opt,name=type,proto3,enum=terminalshooter.ItemType" json:"type,omitempty"`
	Position *Position              `protobuf:"bytes,3,opt,name=position,proto3" json:"position,omitempty"`
	// statusはserverからのみ送信する
	Status        ItemStatus `protobuf:"varint,4,opt,name=status,proto3,enum=terminalshooter.ItemStatus" json:"status,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ItemState) Reset() {
	*x = ItemState{}
	mi := &file_game_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ItemState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ItemState) ProtoMessage() {}

func (x *ItemState) ProtoReflect() protoreflect.Message {
	mi := &file_game_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ItemState.ProtoReflect.Descriptor instead.
func (*ItemState) Descriptor() ([]byte, []int) {
	return file_game_proto_rawDescGZIP(), []int{2}
}

func (x *ItemState) GetItemId() string {
	if x != nil {
		return x.ItemId
	}
	return ""
}

func (x *ItemState) GetType() ItemType {
	if x != nil {
		return x.Type
	}
	return ItemType_BULLET
}

func (x *ItemState) GetPosition() *Position {
	if x != nil {
		return x.Position
	}
	return nil
}

func (x *ItemState) GetStatus() ItemStatus {
	if x != nil {
		return x.Status
	}
	return ItemStatus_ACTIVE
}

var File_game_proto protoreflect.FileDescriptor

var file_game_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x67, 0x61, 0x6d, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x74, 0x65,
	0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x22, 0x26, 0x0a,
	0x08, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x0c, 0x0a, 0x01, 0x78, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x01, 0x79, 0x22, 0xcc, 0x01, 0x0a, 0x0b, 0x50, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x6c, 0x61, 0x79, 0x65, 0x72,
	0x49, 0x64, 0x12, 0x35, 0x0a, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73,
	0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2e, 0x50, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x08, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x38, 0x0a, 0x09, 0x64, 0x69, 0x72,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x74,
	0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2e, 0x44,
	0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x2f, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68,
	0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0xbf, 0x01, 0x0a, 0x09, 0x49, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x69, 0x74, 0x65, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x69, 0x74, 0x65, 0x6d, 0x49, 0x64, 0x12, 0x2d, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x74, 0x65, 0x72, 0x6d,
	0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2e, 0x49, 0x74, 0x65, 0x6d,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x35, 0x0a, 0x08, 0x70, 0x6f,
	0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x74,
	0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2e, 0x50,
	0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x33, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1b, 0x2e, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61, 0x6c, 0x73, 0x68, 0x6f, 0x6f,
	0x74, 0x65, 0x72, 0x2e, 0x49, 0x74, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a, 0x25, 0x0a, 0x0a, 0x49, 0x74, 0x65, 0x6d, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x12, 0x0a, 0x0a, 0x06, 0x41, 0x43, 0x54, 0x49, 0x56, 0x45, 0x10, 0x00,
	0x12, 0x0b, 0x0a, 0x07, 0x52, 0x45, 0x4d, 0x4f, 0x56, 0x45, 0x44, 0x10, 0x01, 0x2a, 0x32, 0x0a,
	0x09, 0x44, 0x69, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x06, 0x0a, 0x02, 0x55, 0x50,
	0x10, 0x00, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04,
	0x4c, 0x45, 0x46, 0x54, 0x10, 0x02, 0x12, 0x09, 0x0a, 0x05, 0x52, 0x49, 0x47, 0x48, 0x54, 0x10,
	0x03, 0x2a, 0x25, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x09, 0x0a, 0x05, 0x41,
	0x4c, 0x49, 0x56, 0x45, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x44, 0x49, 0x53, 0x43, 0x4f, 0x4e,
	0x4e, 0x45, 0x43, 0x54, 0x45, 0x44, 0x10, 0x01, 0x2a, 0x16, 0x0a, 0x08, 0x49, 0x74, 0x65, 0x6d,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x0a, 0x0a, 0x06, 0x42, 0x55, 0x4c, 0x4c, 0x45, 0x54, 0x10, 0x00,
	0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x68, 0x69, 0x62, 0x61, 0x79, 0x75, 0x33, 0x36, 0x2f, 0x74, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x61,
	0x6c, 0x2d, 0x73, 0x68, 0x6f, 0x6f, 0x74, 0x65, 0x72, 0x2f, 0x73, 0x68, 0x61, 0x72, 0x65, 0x64,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_game_proto_rawDescOnce sync.Once
	file_game_proto_rawDescData = file_game_proto_rawDesc
)

func file_game_proto_rawDescGZIP() []byte {
	file_game_proto_rawDescOnce.Do(func() {
		file_game_proto_rawDescData = protoimpl.X.CompressGZIP(file_game_proto_rawDescData)
	})
	return file_game_proto_rawDescData
}

var file_game_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_game_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_game_proto_goTypes = []any{
	(ItemStatus)(0),     // 0: terminalshooter.ItemStatus
	(Direction)(0),      // 1: terminalshooter.Direction
	(Status)(0),         // 2: terminalshooter.Status
	(ItemType)(0),       // 3: terminalshooter.ItemType
	(*Position)(nil),    // 4: terminalshooter.Position
	(*PlayerState)(nil), // 5: terminalshooter.PlayerState
	(*ItemState)(nil),   // 6: terminalshooter.ItemState
}
var file_game_proto_depIdxs = []int32{
	4, // 0: terminalshooter.PlayerState.position:type_name -> terminalshooter.Position
	1, // 1: terminalshooter.PlayerState.direction:type_name -> terminalshooter.Direction
	2, // 2: terminalshooter.PlayerState.status:type_name -> terminalshooter.Status
	3, // 3: terminalshooter.ItemState.type:type_name -> terminalshooter.ItemType
	4, // 4: terminalshooter.ItemState.position:type_name -> terminalshooter.Position
	0, // 5: terminalshooter.ItemState.status:type_name -> terminalshooter.ItemStatus
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_game_proto_init() }
func file_game_proto_init() {
	if File_game_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_game_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_game_proto_goTypes,
		DependencyIndexes: file_game_proto_depIdxs,
		EnumInfos:         file_game_proto_enumTypes,
		MessageInfos:      file_game_proto_msgTypes,
	}.Build()
	File_game_proto = out.File
	file_game_proto_rawDesc = nil
	file_game_proto_goTypes = nil
	file_game_proto_depIdxs = nil
}
