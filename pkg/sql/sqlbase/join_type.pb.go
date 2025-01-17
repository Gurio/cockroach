// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: sql/sqlbase/join_type.proto

package sqlbase

import proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// JoinType is the particular type of a join (or join-like) operation. Not all
// values are used in all contexts.
type JoinType int32

const (
	JoinType_INNER       JoinType = 0
	JoinType_LEFT_OUTER  JoinType = 1
	JoinType_RIGHT_OUTER JoinType = 2
	JoinType_FULL_OUTER  JoinType = 3
	// A left semi join returns the rows from the left side that match at least
	// one row from the right side (as per equality columns and ON condition).
	JoinType_LEFT_SEMI JoinType = 4
	// A left anti join is an "inverted" semi join: it returns the rows from the
	// left side that don't match any columns on the right side (as per equality
	// columns and ON condition).
	JoinType_LEFT_ANTI JoinType = 5
	// INTERSECT_ALL is a special join-like operation that is only used for
	// INTERSECT ALL and INTERSECT operations.
	//
	// It is similar to a left semi join, except that if there are multiple left
	// rows that have the same values on the equality columns, only as many of
	// those are returned as there are matches on the right side.
	//
	// In practice, there is a one-to-one mapping between the left and right
	// columns (they are all equality columns).
	//
	// For example:
	//
	//       Left    Right    Result
	//       1       1        1
	//       1       2        2
	//       2       2        2
	//       2       3        3
	//       3       3
	//               3
	JoinType_INTERSECT_ALL JoinType = 6
	// EXCEPT_ALL is a special join-like operation that is only used for EXCEPT
	// ALL and EXCEPT operations.
	//
	// It is similar to a left anti join, except that if there are multiple left
	// rows that have the same values on the equality columns, only as many of
	// those are removed as there are matches on the right side.
	//
	// In practice, there is a one-to-one mapping between the left and right
	// columns (they are all equality columns).
	//
	// For example:
	//
	//       Left    Right    Result
	//       1       1        1
	//       1       2        2
	//       2       3        2
	//       2       3
	//       2       3
	//       3
	//       3
	//
	//
	// In practice, there is a one-to-one mapping between the left and right
	// columns (they are all equality columns).
	JoinType_EXCEPT_ALL JoinType = 7
)

var JoinType_name = map[int32]string{
	0: "INNER",
	1: "LEFT_OUTER",
	2: "RIGHT_OUTER",
	3: "FULL_OUTER",
	4: "LEFT_SEMI",
	5: "LEFT_ANTI",
	6: "INTERSECT_ALL",
	7: "EXCEPT_ALL",
}
var JoinType_value = map[string]int32{
	"INNER":         0,
	"LEFT_OUTER":    1,
	"RIGHT_OUTER":   2,
	"FULL_OUTER":    3,
	"LEFT_SEMI":     4,
	"LEFT_ANTI":     5,
	"INTERSECT_ALL": 6,
	"EXCEPT_ALL":    7,
}

func (x JoinType) Enum() *JoinType {
	p := new(JoinType)
	*p = x
	return p
}
func (x JoinType) String() string {
	return proto.EnumName(JoinType_name, int32(x))
}
func (x *JoinType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(JoinType_value, data, "JoinType")
	if err != nil {
		return err
	}
	*x = JoinType(value)
	return nil
}
func (JoinType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_join_type_791f1add2f98c606, []int{0}
}

func init() {
	proto.RegisterEnum("cockroach.sql.sqlbase.JoinType", JoinType_name, JoinType_value)
}

func init() {
	proto.RegisterFile("sql/sqlbase/join_type.proto", fileDescriptor_join_type_791f1add2f98c606)
}

var fileDescriptor_join_type_791f1add2f98c606 = []byte{
	// 211 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x2e, 0x2e, 0xcc, 0xd1,
	0x2f, 0x2e, 0xcc, 0x49, 0x4a, 0x2c, 0x4e, 0xd5, 0xcf, 0xca, 0xcf, 0xcc, 0x8b, 0x2f, 0xa9, 0x2c,
	0x48, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0x4d, 0xce, 0x4f, 0xce, 0x2e, 0xca, 0x4f,
	0x4c, 0xce, 0xd0, 0x2b, 0x2e, 0xcc, 0xd1, 0x83, 0x2a, 0xd3, 0x6a, 0x67, 0xe4, 0xe2, 0xf0, 0xca,
	0xcf, 0xcc, 0x0b, 0xa9, 0x2c, 0x48, 0x15, 0xe2, 0xe4, 0x62, 0xf5, 0xf4, 0xf3, 0x73, 0x0d, 0x12,
	0x60, 0x10, 0xe2, 0xe3, 0xe2, 0xf2, 0x71, 0x75, 0x0b, 0x89, 0xf7, 0x0f, 0x0d, 0x71, 0x0d, 0x12,
	0x60, 0x14, 0xe2, 0xe7, 0xe2, 0x0e, 0xf2, 0x74, 0xf7, 0x80, 0x09, 0x30, 0x81, 0x14, 0xb8, 0x85,
	0xfa, 0xf8, 0x40, 0xf9, 0xcc, 0x42, 0xbc, 0x5c, 0x9c, 0x60, 0x0d, 0xc1, 0xae, 0xbe, 0x9e, 0x02,
	0x2c, 0x70, 0xae, 0xa3, 0x5f, 0x88, 0xa7, 0x00, 0xab, 0x90, 0x20, 0x17, 0xaf, 0xa7, 0x5f, 0x88,
	0x6b, 0x50, 0xb0, 0xab, 0x73, 0x48, 0xbc, 0xa3, 0x8f, 0x8f, 0x00, 0x1b, 0xc8, 0x00, 0xd7, 0x08,
	0x67, 0xd7, 0x00, 0x08, 0x9f, 0xdd, 0x49, 0xf1, 0xc4, 0x43, 0x39, 0x86, 0x13, 0x8f, 0xe4, 0x18,
	0x2f, 0x3c, 0x92, 0x63, 0xbc, 0xf1, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5,
	0x18, 0xa2, 0xd8, 0xa1, 0x8e, 0x05, 0x04, 0x00, 0x00, 0xff, 0xff, 0x7d, 0x96, 0x9e, 0x97, 0xe1,
	0x00, 0x00, 0x00,
}
