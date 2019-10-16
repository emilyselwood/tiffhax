package tiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/emilyselwood/tiffhax/parser"
	"github.com/emilyselwood/tiffhax/parser/tiff/constants"
	"github.com/emilyselwood/tiffhax/payload"
	"html/template"
	"io"
)

type Field struct {
	Start int64
	End   int64
	Data  []byte
	ID    uint16
	DType uint16
	Count uint32
	Value interface{}
}

func ParseField(in io.Reader, start int64, order binary.ByteOrder) (*Field, *Offset, error) {
	data := make([]byte, 12)
	n, err := in.Read(data)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read ifd field, %v", err)
	}
	if n != 12 {
		return nil, nil, fmt.Errorf("strange size read from ifd field got %v expected 12", n)
	}

	var result Field
	result.Start = start
	result.End = start + 12
	result.Data = data

	// parse the actual data.
	result.ID = order.Uint16(data[0:2])
	result.DType = order.Uint16(data[2:4])
	result.Count = order.Uint32(data[4:8])
	result.Value = order.Uint32(data[8:12])

	// do we have an offset or a value
	if result.Count * constants.DataTypeSize[result.DType] > 4 {
		var offset Offset
		offset.DType = int(result.DType)
		offset.From = start
		offset.To = int64(order.Uint32(data[8:12]))
		offset.Count = result.Count

		return &result, &offset, nil
	}

	return &result, nil, nil
}


func (f *Field) Contains(offset int64) bool {
	return f.Start <= offset && offset < f.End
}

func (f *Field) ContainsRegion(start int64, end int64) bool {
	return f.Start <= start && start < f.End && f.Start < end && end < f.End
}

func (f *Field) Find(offset int64) (parser.Region, error) {
	if offset < f.Start || offset >= f.End {
		return nil, fmt.Errorf("find offset %v outside of field region %v to %v", offset, f.Start, f.End)
	}
	return f, nil
}

func (f *Field) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("field can not be split")
}

func (f *Field) Render() ([]payload.Section, error) {

	desc, err := payload.RenderTemplate(fieldTemplate, f, template.FuncMap{
		"FieldNames": func(fieldId uint16) string {
			return constants.FieldNames[fieldId]
		},
		"DataTypeNames" : func(typeId uint16) string {
			return constants.DataTypeNames[typeId]
		},
		// TODO: lookups for field value meanings
		// TODO: better descriptions
	})
	if err != nil {
		return nil, fmt.Errorf("could not format template for field, %v", err)
	}

	var data bytes.Buffer

	payload.RenderBytesSpan(&data, f.Data[0:2], "field_id")
	data.WriteRune(' ')
	payload.RenderBytesSpan(&data, f.Data[2:4], "field_type")
	data.WriteRune(' ')
	payload.RenderBytesSpan(&data, f.Data[4:8], "field_count")
	data.WriteRune(' ')
	payload.RenderBytesSpan(&data, f.Data[8:12], "field_value")

	return []payload.Section{
		&payload.General{
			Start:   f.Start,
			End:     f.End - 1,
			Id:      "ifd_field",
			TheData: template.HTML(data.String()),
			Text:    template.HTML(desc),
		},
	}, nil

}

const fieldTemplate = `A field called <span class="field_id">{{ FieldNames .ID}}</span> is a <span class="field_type">{{ DataTypeNames .DType }}</span> with <span class="field_count">{{.Count}}</span> entries and value <span class="field_value">{{.Value}}</span>`