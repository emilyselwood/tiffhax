package tiff

import (
	"encoding/binary"
	"fmt"
	"github.com/emilyselwood/tiffhax/parser"
	"github.com/emilyselwood/tiffhax/parser/tiff/constants"
	"github.com/emilyselwood/tiffhax/payload"
	"html/template"
	"io"
)

/*
Offset holds a link from somewhere in a file to somewhere else.
*/
type Offset struct {
	From    int64
	To      int64
	Start   int64
	End     int64
	DType   uint16
	Count   uint32
	FieldId uint16
	IsData  bool
	Data    []byte
	IFD     *IFD
}


func (o *Offset) Parse(in io.ReadSeeker, order binary.ByteOrder) ([]*Data, error) {
	_, err := in.Seek(o.To, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("could not seek to offset start %v, %v", o.To, err)
	}

	o.Start = o.To
	o.End = o.Start + (int64(o.Count) * int64(constants.DataTypeSize[o.DType]))
	if !o.IsData {
		o.Data = make([]byte, o.End - o.Start)
		n, err := in.Read(o.Data)
		if err != nil {
			return nil, fmt.Errorf("could not read data at offset %v, %v", o.To, err)
		}
		if int64(n) != o.End - o.Start {
			return nil, fmt.Errorf("did not get enough data when reading at offset %v", o.To)
		}
		return nil, nil
	}
	chunk := make([]byte, constants.DataTypeSize[o.DType])
	var data []*Data
	for i := 0; uint32(i) < o.Count; i++ {
		var d Data
		n, err := in.Read(chunk)
		if err != nil {
			return nil, fmt.Errorf("could not read offset entry, %v", err)
		}
		if uint32(n) != constants.DataTypeSize[o.DType] {
			return nil, fmt.Errorf("got wrong number of bytes from read, expected %v got %v", constants.DataTypeSize[o.DType], n)
		}

		o.Data = append(o.Data, chunk...)

		d.Start = int64(order.Uint32(chunk))
		d.IFD = o.IFD
		d.I = i
		data = append(data, &d)
	}

	return data, nil
}


func (o *Offset) Contains(offset int64) bool {
	return o.Start <= offset && offset < o.End
}

func (o *Offset) ContainsRegion(start int64, end int64) bool {
	return o.Start <= start && start < o.End && o.Start < end && end < o.End
}

func (o *Offset) Find(offset int64) (parser.Region, error) {
	if offset < o.Start || offset >= o.End {
		return nil, fmt.Errorf("find offset %v outside of offset region %v to %v", offset, o.Start, o.End)
	}
	return o, nil
}

func (o *Offset) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("offset at %v to %v can not be split between %v and %v", o.Start, o.End, start, end)
}

func (o *Offset) Render() ([]payload.Section, error) {

	desc, err := payload.RenderTemplate(offsetTemplate, o, template.FuncMap{
		"FieldNames": func(fieldId uint16) string {
			return constants.FieldNames[fieldId]
		},
		"DataTypeNames": func(typeId uint16) string {
			return constants.DataTypeNames[typeId]
		},
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't render offset description, %v", err)
	}

	var data string
	if len(o.Data) > 0 {
		data = payload.RenderBytes(o.Data)
	}

	return []payload.Section{
		&payload.General{
			Start:   o.Start,
			End:     o.End - 1,
			Id:      "offset",
			TheData: template.HTML(data),
			Text:    template.HTML(desc),
		},
	}, nil
}

const offsetTemplate = `{{.Count}} {{DataTypeNames .DType}} values for <a href="#{{.From}}">{{FieldNames .FieldId}}</a>`