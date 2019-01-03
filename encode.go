package teafile

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
	"strings"
	"time"
)

type Encoder struct {
	byteOrder binary.ByteOrder
	w         *bufio.Writer

	itemName string
}

func Encode(writer io.Writer, header *Header, items interface{}) error {
	return nil
}

func NewEncoder(w io.Writer, header *Header) (*Encoder, error) {
	e := &Encoder{byteOrder: binary.BigEndian, w: bufio.NewWriter(w)}

	e.binaryWrite(MagicValueBigEndian)

	return e, nil
}

func (enc *Encoder) binaryWrite(value interface{}) error {
	return binary.Write(enc.w, enc.byteOrder, value)
}

func (enc *Encoder) Encode(item interface{}) error {
	if err := enc.marshal(item); err != nil {
		return err
	}

	return enc.w.Flush()
}

func (enc *Encoder) marshal(v interface{}) error {
	val := reflect.ValueOf(v)

	if !val.IsValid() {
		return nil
	}

	for val.Kind() == reflect.Interface || val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	k := val.Kind()
	t := val.Type()

	if enc.itemName == "" {
		enc.itemName = t.String()
		// enc.binaryWrite(itemSection)
		enc.binaryWrite([]byte(enc.itemName))
	}

	if k != reflect.Struct {
		return errors.New("")
	}

	if t.NumField() == 0 {
		return errors.New("")
	}

	var foundTSTag bool
	for idx := 0; idx < t.NumField(); idx++ {
		sf := t.Field(idx)

		if tf, ok := sf.Tag.Lookup(StructTagName); ok {
			if strings.HasSuffix(tf, ","+StructTagTimestamp) {
				foundTSTag = true
			}
		}

		fval := val.FieldByName(sf.Name)
		var value interface{}

		switch v := fval.Interface().(type) {
		case int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
			value = v
		case time.Time:
			value = v.UnixNano()
		}

		enc.binaryWrite(value)
	}
	if !foundTSTag {
		return errors.New("")
	}

	return nil
}
