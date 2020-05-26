// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: proto/types/params.proto

package types

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/duration"
	math "math"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// ConsensusParams contains consensus critical parameters that determine the
// validity of blocks.
type ConsensusParams struct {
	Block                BlockParams     `protobuf:"bytes,1,opt,name=block,proto3" json:"block"`
	Evidence             EvidenceParams  `protobuf:"bytes,2,opt,name=evidence,proto3" json:"evidence"`
	Validator            ValidatorParams `protobuf:"bytes,3,opt,name=validator,proto3" json:"validator"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ConsensusParams) Reset()         { *m = ConsensusParams{} }
func (m *ConsensusParams) String() string { return proto.CompactTextString(m) }
func (*ConsensusParams) ProtoMessage()    {}
func (*ConsensusParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_95a9f934fa6f056c, []int{0}
}
func (m *ConsensusParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsensusParams.Unmarshal(m, b)
}
func (m *ConsensusParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsensusParams.Marshal(b, m, deterministic)
}
func (m *ConsensusParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsensusParams.Merge(m, src)
}
func (m *ConsensusParams) XXX_Size() int {
	return xxx_messageInfo_ConsensusParams.Size(m)
}
func (m *ConsensusParams) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsensusParams.DiscardUnknown(m)
}

var xxx_messageInfo_ConsensusParams proto.InternalMessageInfo

func (m *ConsensusParams) GetBlock() BlockParams {
	if m != nil {
		return m.Block
	}
	return BlockParams{}
}

func (m *ConsensusParams) GetEvidence() EvidenceParams {
	if m != nil {
		return m.Evidence
	}
	return EvidenceParams{}
}

func (m *ConsensusParams) GetValidator() ValidatorParams {
	if m != nil {
		return m.Validator
	}
	return ValidatorParams{}
}

// BlockParams contains limits on the block size.
type BlockParams struct {
	// Note: must be greater than 0
	MaxBytes int64 `protobuf:"varint,1,opt,name=max_bytes,json=maxBytes,proto3" json:"max_bytes,omitempty"`
	// Note: must be greater or equal to -1
	MaxGas int64 `protobuf:"varint,2,opt,name=max_gas,json=maxGas,proto3" json:"max_gas,omitempty"`
	// Minimum time increment between consecutive blocks (in milliseconds)
	// Not exposed to the application.
	TimeIotaMs           int64    `protobuf:"varint,3,opt,name=time_iota_ms,json=timeIotaMs,proto3" json:"time_iota_ms,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockParams) Reset()         { *m = BlockParams{} }
func (m *BlockParams) String() string { return proto.CompactTextString(m) }
func (*BlockParams) ProtoMessage()    {}
func (*BlockParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_95a9f934fa6f056c, []int{1}
}
func (m *BlockParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockParams.Unmarshal(m, b)
}
func (m *BlockParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockParams.Marshal(b, m, deterministic)
}
func (m *BlockParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockParams.Merge(m, src)
}
func (m *BlockParams) XXX_Size() int {
	return xxx_messageInfo_BlockParams.Size(m)
}
func (m *BlockParams) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockParams.DiscardUnknown(m)
}

var xxx_messageInfo_BlockParams proto.InternalMessageInfo

func (m *BlockParams) GetMaxBytes() int64 {
	if m != nil {
		return m.MaxBytes
	}
	return 0
}

func (m *BlockParams) GetMaxGas() int64 {
	if m != nil {
		return m.MaxGas
	}
	return 0
}

func (m *BlockParams) GetTimeIotaMs() int64 {
	if m != nil {
		return m.TimeIotaMs
	}
	return 0
}

// EvidenceParams determine how we handle evidence of malfeasance.
type EvidenceParams struct {
	// Max age of evidence, in blocks.
	//
	// The basic formula for calculating this is: MaxAgeDuration / {average block
	// time}.
	MaxAgeNumBlocks int64 `protobuf:"varint,1,opt,name=max_age_num_blocks,json=maxAgeNumBlocks,proto3" json:"max_age_num_blocks,omitempty"`
	// Max age of evidence, in time.
	//
	// It should correspond with an app's "unbonding period" or other similar
	// mechanism for handling [Nothing-At-Stake
	// attacks](https://github.com/ethereum/wiki/wiki/Proof-of-Stake-FAQ#what-is-the-nothing-at-stake-problem-and-how-can-it-be-fixed).
	MaxAgeDuration time.Duration `protobuf:"bytes,2,opt,name=max_age_duration,json=maxAgeDuration,proto3,stdduration" json:"max_age_duration"`
	// This sets the maximum number of evidence that can be committed in a single block.
	// and should fall comfortably under the max block bytes when we consider the size of
	// each evidence (See MaxEvidenceBytes). The maximum number is MaxEvidencePerBlock.
	// Default is 50
	MaxNum               uint32   `protobuf:"varint,3,opt,name=max_num,json=maxNum,proto3" json:"max_num,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EvidenceParams) Reset()         { *m = EvidenceParams{} }
func (m *EvidenceParams) String() string { return proto.CompactTextString(m) }
func (*EvidenceParams) ProtoMessage()    {}
func (*EvidenceParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_95a9f934fa6f056c, []int{2}
}
func (m *EvidenceParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EvidenceParams.Unmarshal(m, b)
}
func (m *EvidenceParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EvidenceParams.Marshal(b, m, deterministic)
}
func (m *EvidenceParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EvidenceParams.Merge(m, src)
}
func (m *EvidenceParams) XXX_Size() int {
	return xxx_messageInfo_EvidenceParams.Size(m)
}
func (m *EvidenceParams) XXX_DiscardUnknown() {
	xxx_messageInfo_EvidenceParams.DiscardUnknown(m)
}

var xxx_messageInfo_EvidenceParams proto.InternalMessageInfo

func (m *EvidenceParams) GetMaxAgeNumBlocks() int64 {
	if m != nil {
		return m.MaxAgeNumBlocks
	}
	return 0
}

func (m *EvidenceParams) GetMaxAgeDuration() time.Duration {
	if m != nil {
		return m.MaxAgeDuration
	}
	return 0
}

func (m *EvidenceParams) GetMaxNum() uint32 {
	if m != nil {
		return m.MaxNum
	}
	return 0
}

// ValidatorParams restrict the public key types validators can use.
// NOTE: uses ABCI pubkey naming, not Amino names.
type ValidatorParams struct {
	PubKeyTypes          []string `protobuf:"bytes,1,rep,name=pub_key_types,json=pubKeyTypes,proto3" json:"pub_key_types,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ValidatorParams) Reset()         { *m = ValidatorParams{} }
func (m *ValidatorParams) String() string { return proto.CompactTextString(m) }
func (*ValidatorParams) ProtoMessage()    {}
func (*ValidatorParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_95a9f934fa6f056c, []int{3}
}
func (m *ValidatorParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ValidatorParams.Unmarshal(m, b)
}
func (m *ValidatorParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ValidatorParams.Marshal(b, m, deterministic)
}
func (m *ValidatorParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ValidatorParams.Merge(m, src)
}
func (m *ValidatorParams) XXX_Size() int {
	return xxx_messageInfo_ValidatorParams.Size(m)
}
func (m *ValidatorParams) XXX_DiscardUnknown() {
	xxx_messageInfo_ValidatorParams.DiscardUnknown(m)
}

var xxx_messageInfo_ValidatorParams proto.InternalMessageInfo

func (m *ValidatorParams) GetPubKeyTypes() []string {
	if m != nil {
		return m.PubKeyTypes
	}
	return nil
}

// HashedParams is a subset of ConsensusParams.
// It is amino encoded and hashed into
// the Header.ConsensusHash.
type HashedParams struct {
	BlockMaxBytes        int64    `protobuf:"varint,1,opt,name=block_max_bytes,json=blockMaxBytes,proto3" json:"block_max_bytes,omitempty"`
	BlockMaxGas          int64    `protobuf:"varint,2,opt,name=block_max_gas,json=blockMaxGas,proto3" json:"block_max_gas,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashedParams) Reset()         { *m = HashedParams{} }
func (m *HashedParams) String() string { return proto.CompactTextString(m) }
func (*HashedParams) ProtoMessage()    {}
func (*HashedParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_95a9f934fa6f056c, []int{4}
}
func (m *HashedParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashedParams.Unmarshal(m, b)
}
func (m *HashedParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashedParams.Marshal(b, m, deterministic)
}
func (m *HashedParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashedParams.Merge(m, src)
}
func (m *HashedParams) XXX_Size() int {
	return xxx_messageInfo_HashedParams.Size(m)
}
func (m *HashedParams) XXX_DiscardUnknown() {
	xxx_messageInfo_HashedParams.DiscardUnknown(m)
}

var xxx_messageInfo_HashedParams proto.InternalMessageInfo

func (m *HashedParams) GetBlockMaxBytes() int64 {
	if m != nil {
		return m.BlockMaxBytes
	}
	return 0
}

func (m *HashedParams) GetBlockMaxGas() int64 {
	if m != nil {
		return m.BlockMaxGas
	}
	return 0
}

func init() {
	proto.RegisterType((*ConsensusParams)(nil), "tendermint.proto.types.ConsensusParams")
	proto.RegisterType((*BlockParams)(nil), "tendermint.proto.types.BlockParams")
	proto.RegisterType((*EvidenceParams)(nil), "tendermint.proto.types.EvidenceParams")
	proto.RegisterType((*ValidatorParams)(nil), "tendermint.proto.types.ValidatorParams")
	proto.RegisterType((*HashedParams)(nil), "tendermint.proto.types.HashedParams")
}

func init() { proto.RegisterFile("proto/types/params.proto", fileDescriptor_95a9f934fa6f056c) }

var fileDescriptor_95a9f934fa6f056c = []byte{
	// 469 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x53, 0xd1, 0x6a, 0xd4, 0x40,
	0x14, 0x35, 0xae, 0xd6, 0xdd, 0xbb, 0xdd, 0xae, 0xcc, 0x83, 0xc6, 0x0a, 0xed, 0x12, 0x61, 0x2d,
	0x28, 0x09, 0x54, 0x7c, 0x16, 0xa3, 0xd2, 0x4a, 0xd9, 0x22, 0x41, 0x7c, 0xe8, 0xcb, 0x70, 0xb3,
	0x19, 0xb3, 0xa1, 0x3b, 0x99, 0x90, 0x99, 0x29, 0x9b, 0x3f, 0xf1, 0x07, 0x04, 0x3f, 0xc5, 0xaf,
	0x50, 0xf0, 0xcd, 0xbf, 0x90, 0xcc, 0xec, 0x98, 0xdd, 0xd2, 0xbe, 0xcd, 0xdc, 0x7b, 0xce, 0x99,
	0x7b, 0xce, 0x65, 0xc0, 0xaf, 0x6a, 0xa1, 0x44, 0xa4, 0x9a, 0x8a, 0xc9, 0xa8, 0xc2, 0x1a, 0xb9,
	0x0c, 0x4d, 0x89, 0x3c, 0x52, 0xac, 0xcc, 0x58, 0xcd, 0x8b, 0x52, 0xd9, 0x4a, 0x68, 0x40, 0xfb,
	0x53, 0xb5, 0x28, 0xea, 0x8c, 0x56, 0x58, 0xab, 0x26, 0xb2, 0xec, 0x5c, 0xe4, 0xa2, 0x3b, 0x59,
	0xf4, 0xfe, 0x41, 0x2e, 0x44, 0xbe, 0x64, 0x16, 0x92, 0xea, 0xaf, 0x51, 0xa6, 0x6b, 0x54, 0x85,
	0x28, 0x6d, 0x3f, 0xf8, 0xeb, 0xc1, 0xf8, 0x9d, 0x28, 0x25, 0x2b, 0xa5, 0x96, 0x9f, 0xcc, 0xcb,
	0xe4, 0x0d, 0xdc, 0x4f, 0x97, 0x62, 0x7e, 0xe9, 0x7b, 0x13, 0xef, 0x68, 0x78, 0xfc, 0x2c, 0xbc,
	0x79, 0x86, 0x30, 0x6e, 0x41, 0x96, 0x13, 0xdf, 0xfb, 0xf9, 0xeb, 0xf0, 0x4e, 0x62, 0x79, 0xe4,
	0x14, 0xfa, 0xec, 0xaa, 0xc8, 0x58, 0x39, 0x67, 0xfe, 0x5d, 0xa3, 0x31, 0xbd, 0x4d, 0xe3, 0xc3,
	0x1a, 0xb7, 0x25, 0xf3, 0x9f, 0x4d, 0xce, 0x60, 0x70, 0x85, 0xcb, 0x22, 0x43, 0x25, 0x6a, 0xbf,
	0x67, 0xa4, 0x9e, 0xdf, 0x26, 0xf5, 0xc5, 0x01, 0xb7, 0xb4, 0x3a, 0x7e, 0xc0, 0x60, 0xb8, 0x31,
	0x32, 0x79, 0x0a, 0x03, 0x8e, 0x2b, 0x9a, 0x36, 0x8a, 0x49, 0x63, 0xb5, 0x97, 0xf4, 0x39, 0xae,
	0xe2, 0xf6, 0x4e, 0x1e, 0xc3, 0x83, 0xb6, 0x99, 0xa3, 0x34, 0x0e, 0x7a, 0xc9, 0x0e, 0xc7, 0xd5,
	0x09, 0x4a, 0x32, 0x81, 0x5d, 0x55, 0x70, 0x46, 0x0b, 0xa1, 0x90, 0x72, 0x69, 0x86, 0xea, 0x25,
	0xd0, 0xd6, 0x3e, 0x0a, 0x85, 0x33, 0x19, 0x7c, 0xf7, 0x60, 0x6f, 0xdb, 0x16, 0x79, 0x01, 0xa4,
	0x55, 0xc3, 0x9c, 0xd1, 0x52, 0x73, 0x6a, 0x52, 0x72, 0x6f, 0x8e, 0x39, 0xae, 0xde, 0xe6, 0xec,
	0x5c, 0x73, 0x33, 0x9c, 0x24, 0x33, 0x78, 0xe8, 0xc0, 0x6e, 0x59, 0xeb, 0x14, 0x9f, 0x84, 0x76,
	0x9b, 0xa1, 0xdb, 0x66, 0xf8, 0x7e, 0x0d, 0x88, 0xfb, 0xad, 0xd9, 0x6f, 0xbf, 0x0f, 0xbd, 0x64,
	0xcf, 0xea, 0xb9, 0x8e, 0x73, 0x52, 0x6a, 0x6e, 0x66, 0x1d, 0x19, 0x27, 0xe7, 0x9a, 0x07, 0xaf,
	0x61, 0x7c, 0x2d, 0x32, 0x12, 0xc0, 0xa8, 0xd2, 0x29, 0xbd, 0x64, 0x0d, 0x35, 0x99, 0xfa, 0xde,
	0xa4, 0x77, 0x34, 0x48, 0x86, 0x95, 0x4e, 0xcf, 0x58, 0xf3, 0xb9, 0x2d, 0x05, 0x17, 0xb0, 0x7b,
	0x8a, 0x72, 0xc1, 0xb2, 0x35, 0x67, 0x0a, 0x63, 0xe3, 0x87, 0x5e, 0x0f, 0x73, 0x64, 0xca, 0x33,
	0x97, 0x68, 0x00, 0xa3, 0x0e, 0xd7, 0xe5, 0x3a, 0x74, 0xa8, 0x13, 0x94, 0xf1, 0xf1, 0x8f, 0x3f,
	0x07, 0xde, 0xc5, 0xcb, 0xbc, 0x50, 0x0b, 0x9d, 0x86, 0x73, 0xc1, 0xa3, 0x6e, 0xd7, 0x9b, 0xc7,
	0x8d, 0xef, 0x92, 0xee, 0x98, 0xcb, 0xab, 0x7f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x8a, 0x3c, 0x66,
	0x44, 0x44, 0x03, 0x00, 0x00,
}

func (this *ConsensusParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ConsensusParams)
	if !ok {
		that2, ok := that.(ConsensusParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Block.Equal(&that1.Block) {
		return false
	}
	if !this.Evidence.Equal(&that1.Evidence) {
		return false
	}
	if !this.Validator.Equal(&that1.Validator) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *BlockParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*BlockParams)
	if !ok {
		that2, ok := that.(BlockParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxBytes != that1.MaxBytes {
		return false
	}
	if this.MaxGas != that1.MaxGas {
		return false
	}
	if this.TimeIotaMs != that1.TimeIotaMs {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *EvidenceParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvidenceParams)
	if !ok {
		that2, ok := that.(EvidenceParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxAgeNumBlocks != that1.MaxAgeNumBlocks {
		return false
	}
	if this.MaxAgeDuration != that1.MaxAgeDuration {
		return false
	}
	if this.MaxNum != that1.MaxNum {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ValidatorParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ValidatorParams)
	if !ok {
		that2, ok := that.(ValidatorParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.PubKeyTypes) != len(that1.PubKeyTypes) {
		return false
	}
	for i := range this.PubKeyTypes {
		if this.PubKeyTypes[i] != that1.PubKeyTypes[i] {
			return false
		}
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *HashedParams) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*HashedParams)
	if !ok {
		that2, ok := that.(HashedParams)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.BlockMaxBytes != that1.BlockMaxBytes {
		return false
	}
	if this.BlockMaxGas != that1.BlockMaxGas {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}