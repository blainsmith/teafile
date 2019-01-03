package teafile

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
)

// A Decoder reads and decodes binary teafile data from an input stream
type Decoder struct {
	reader io.Reader

	itemStart int64
	itemEnd   int64

	sectionCount int64

	header *Header

	itemSection      itemSection
	contentSection   contentSection
	nameValueSection nameValueSection
	timeSection      timeSection

	padding []byte
}

// Decode reads and decode all of the binary data from an input stream into a Header and items. The rows should be a slice to hold all of the rows from the input stream.
func Decode(reader io.Reader, header *Header, rows interface{}) error {
	if rows == nil {
		return ErrNonPointer
	}

	rv := reflect.ValueOf(rows)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrNonPointer
	}

	e := rv.Elem()
	if e.Kind() != reflect.Slice {
		return &UnsupportedTypeError{e.Type()}
	}

	_, err := NewDecoder(reader, header)
	if err != nil {
		return err
	}

	return nil
}

// NewDecoder returns a new decoder that reads from an input stream. It will also decode the header information into the provided Header.
func NewDecoder(reader io.Reader, header *Header) (*Decoder, error) {
	d := Decoder{
		reader: reader,
		header: header,
	}

	if d.header == nil {
		d.header = &DefaultHeader
	}

	if d.header.ByteOrder == nil {
		d.header.ByteOrder = binary.LittleEndian
	}

	if err := d.readEndianness(); err != nil {
		return nil, err
	}
	if err := d.readItemsStartEnd(); err != nil {
		return nil, err
	}
	if err := d.readSectionCount(); err != nil {
		return nil, err
	}
	if err := d.readSections(); err != nil {
		return nil, err
	}
	if err := d.readPadding(); err != nil {
		return nil, err
	}

	return &d, nil
}

// Decode reads and decodes the next row from the input stream. If there is nothing left to decode then an io.EOF error is returned.
func (dec *Decoder) Decode(row interface{}) error {
	return nil
}

func (dec *Decoder) readEndianness() error {
	var mv int64
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &mv); err != nil {
		return fmt.Errorf("reading magic value: %v", err)
	}

	switch mv {
	case MagicValueLittleEndian:
		dec.header.ByteOrder = binary.LittleEndian
	case MagicValueBigEndian:
		dec.header.ByteOrder = binary.BigEndian
	default:
		return errors.New("not a teafile")
	}

	return nil
}

func (dec *Decoder) readItemsStartEnd() error {
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.itemStart); err != nil {
		return fmt.Errorf("reading item start: %v", err)
	}
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.itemEnd); err != nil {
		return fmt.Errorf("reading item end: %v", err)
	}

	return nil
}

func (dec *Decoder) readSectionCount() error {
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.sectionCount); err != nil {
		return fmt.Errorf("reading section count: %v", err)
	}

	return nil
}

func (dec *Decoder) readSections() error {
	var id SecionID
	var err error

	for i := 1; i <= int(dec.sectionCount); i++ {
		if err = binary.Read(dec.reader, dec.header.ByteOrder, &id); err != nil {
			return fmt.Errorf("reading section id: %v", err)
		}

		switch id {
		case SectionIDItem:
			err = dec.readItemSection()
		case SectionIDContent:
			err = dec.readContentSection()
		case SectionIDNameValue:
			err = dec.readNameValueSection()
		case SectionIDTime:
			err = dec.readTimeSection()
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (dec *Decoder) readItemSection() error {
	var nameLen int32
	var name []byte
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.itemSection.nextOffset); err != nil {
		return fmt.Errorf("item section next offset: %v", err)
	}

	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.itemSection.itemSize); err != nil {
		return fmt.Errorf("item section item size: %v", err)
	}

	if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameLen); err != nil {
		return fmt.Errorf("item section name length: %v", err)
	}

	name = make([]byte, nameLen)
	if err := binary.Read(dec.reader, dec.header.ByteOrder, name); err != nil {
		return fmt.Errorf("item section name: %v", err)
	}

	dec.header.Name = string(name)

	if err := dec.readFields(); err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) readFields() error {
	var fieldsCount int32
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &fieldsCount); err != nil {
		return fmt.Errorf("item section fields: %v", err)
	}

	var fieldType FieldType
	var offset int32
	var nameLen int32
	var name []byte
	dec.header.Fields = make([]Field, fieldsCount)
	for idx := range dec.header.Fields {
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &fieldType); err != nil {
			return fmt.Errorf("item section field id %d: %v", idx, err)
		}
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &offset); err != nil {
			return fmt.Errorf("item section field offset %d: %v", idx, err)
		}
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameLen); err != nil {
			return fmt.Errorf("item section field name length %d: %v", idx, err)
		}

		name = make([]byte, nameLen)
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &name); err != nil {
			return fmt.Errorf("item section field name %d: %v", idx, err)
		}

		dec.header.Fields[idx] = Field{
			Name:   string(name),
			Type:   fieldType,
			Offset: int(offset),
		}
	}

	return nil
}

func (dec *Decoder) readContentSection() error {
	var nameLen int32
	var name []byte

	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.contentSection.nextOffset); err != nil {
		return fmt.Errorf("item section next offset: %v", err)
	}
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameLen); err != nil {
		return fmt.Errorf("item section name length: %v", err)
	}

	name = make([]byte, nameLen)
	if err := binary.Read(dec.reader, dec.header.ByteOrder, name); err != nil {
		return fmt.Errorf("item section name: %v", err)
	}

	dec.header.Description = string(name)

	return nil
}

func (dec *Decoder) readNameValueSection() error {
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.nameValueSection.nextOffset); err != nil {
		return fmt.Errorf("item section next offset: %v", err)
	}

	if err := dec.readNameValues(); err != nil {
		return err
	}

	return nil
}

func (dec *Decoder) readNameValues() error {
	var nameValuesCount int32
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameValuesCount); err != nil {
		return fmt.Errorf("item section item size: %v", err)
	}

	var nameLen int32
	var name []byte
	var kind NameValueKind
	dec.header.NameValues = make([]NameValue, nameValuesCount)
	for idx := range dec.nameValueSection.nameValues {
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameLen); err != nil {
			return fmt.Errorf("item section name length: %v", err)
		}

		name = make([]byte, nameLen)
		if err := binary.Read(dec.reader, dec.header.ByteOrder, name); err != nil {
			return fmt.Errorf("item section name: %v", err)
		}
		dec.header.NameValues[idx].Name = string(name)

		if err := binary.Read(dec.reader, dec.header.ByteOrder, &kind); err != nil {
			return fmt.Errorf("name values section name value kind %d: %v", idx, err)
		}
		dec.header.NameValues[idx].Kind = kind

		switch kind {
		case 1:
			var i32 int32
			if err := binary.Read(dec.reader, dec.header.ByteOrder, &i32); err != nil {
				return fmt.Errorf("name values section name value int32 %d: %v", idx, err)
			}
			dec.header.NameValues[idx].Value = i32
		case 2:
			var f64 float64
			if err := binary.Read(dec.reader, dec.header.ByteOrder, &f64); err != nil {
				return fmt.Errorf("name values section name value float64 %d: %v", idx, err)
			}
			dec.header.NameValues[idx].Value = f64
		case 3:
			var nameLen int32
			if err := binary.Read(dec.reader, dec.header.ByteOrder, &nameLen); err != nil {
				return fmt.Errorf("name values section name value string %d: %v", idx, err)
			}

			name := make([]byte, nameLen)
			if err := binary.Read(dec.reader, dec.header.ByteOrder, name); err != nil {
				return fmt.Errorf("name values section name value string %d: %v", idx, err)
			}

			dec.header.NameValues[idx].Value = name
		}

	}

	return nil
}

func (dec *Decoder) readTimeSection() error {
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.timeSection.nextOffset); err != nil {
		return fmt.Errorf("time section next offset: %v", err)
	}
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.timeSection.epoch); err != nil {
		return fmt.Errorf("time section epoch: %v", err)
	}
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.timeSection.ticksPerDay); err != nil {
		return fmt.Errorf("time section ticks per day: %v", err)
	}
	if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.timeSection.fieldsCount); err != nil {
		return fmt.Errorf("time section ticks per day: %v", err)
	}

	dec.timeSection.fieldOffsets = make([]int32, dec.timeSection.fieldsCount)
	for idx := range dec.timeSection.fieldOffsets {
		if err := binary.Read(dec.reader, dec.header.ByteOrder, &dec.timeSection.fieldOffsets[idx]); err != nil {
			return fmt.Errorf("time section field offset %d: %v", idx, err)
		}
	}

	return nil
}

func (dec *Decoder) readPadding() error {
	offset := dec.itemStart

	if dec.itemSection.nextOffset > 0 {
		offset -= 36 + 4 + int64(dec.itemSection.nextOffset)
	}
	if dec.contentSection.nextOffset > 0 {
		offset -= 8 + int64(dec.contentSection.nextOffset)
	}
	if dec.nameValueSection.nextOffset > 0 {
		offset -= 8 + int64(dec.nameValueSection.nextOffset)
	}
	if dec.timeSection.nextOffset > 0 {
		offset -= 8 + int64(dec.timeSection.nextOffset)
	}

	dec.padding = make([]byte, offset)
	if err := binary.Read(dec.reader, dec.header.ByteOrder, dec.padding); err != nil {
		return fmt.Errorf("padding: %v", err)
	}

	return nil
}
