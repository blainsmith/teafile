package teafile

import (
	"encoding/binary"
	"errors"
	"reflect"
)

type SecionID int32
type FieldType int32
type NameValueKind int32

const (
	MagicValueLittleEndian int64 = 0x0d0e0a0402080500
	MagicValueBigEndian    int64 = 0x00050802040a0e0d
	StructTagName                = "teafiles"
	StructTagTimestamp           = "timestamp"

	SectionIDItem      SecionID = 0x0a
	SectionIDTime      SecionID = 0x40
	SectionIDContent   SecionID = 0x80
	SectionIDNameValue SecionID = 0x81

	FieldTypeInt8 FieldType = iota + 1
	FieldTypeInt16
	FieldTypeInt32
	FieldTypeInt64
	FieldTypeUint8
	FieldTypeUint16
	FieldTypeUint32
	FieldTypeUint64
	FieldTypeFloat
	FieldTypeDouble

	NameValueKindInt32 NameValueKind = iota + 1
	NameValueKindDouble
	NameValueKindText
	NameValueKindUUID

	Epoch       = 719162
	TicksPerDay = 86400000
)

var (
	ErrNonPointer = errors.New("non-pointer passed")
)

// An UnsupportedTypeError is returned when attempting to decode or encode an unsupported value type.
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "teafiles: unsupported type: " + e.Type.String()
}

// An UnsupportedValueError is returned when attempting to decode or encode an unsupported value type.
type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "teafiles: unsupported value: " + e.Str
}

type itemSection struct {
	nextOffset int32
	itemSize   int32
}

type Field struct {
	Name   string
	Type   FieldType
	Offset int
}

type contentSection struct {
	nextOffset int32
}

type nameValueSection struct {
	nextOffset      int32
	nameValuesCount int32
	nameValues      []NameValue
}

type NameValue struct {
	Kind  NameValueKind
	Name  string
	Value interface{}
}

type timeSection struct {
	nextOffset   int32
	epoch        int64
	ticksPerDay  int64
	fieldsCount  int32
	fieldOffsets []int32
}

type Times struct {
	Epoch       int64
	TicksPerDay int64
	Offsets     []int32
}

type Header struct {
	ByteOrder   binary.ByteOrder
	Name        string
	Description string
	Fields      []Field
	NameValues  []NameValue
	Times       Times
}

var DefaultHeader = Header{
	ByteOrder: binary.LittleEndian,
	Times: Times{
		Epoch:       Epoch,
		TicksPerDay: TicksPerDay,
	},
}
