package tiff

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/emilyselwood/tiffhax/parser"
	"github.com/emilyselwood/tiffhax/payload"
	"html/template"
	"io"
)


type Header struct {
	Start          int64
	End            int64
	Data           []byte
	Endian         binary.ByteOrder
	BigTiff        bool
	FirstIFDOffset int64
}

func ParseHeader(in io.Reader) (*Header, int64, error) {
	data := make([]byte, 8)

	n, err := in.Read(data)
	if err != nil {
		return nil, int64(n), fmt.Errorf("could not read header, %v", err)
	}
	if n != 8 {
		return nil, int64(n), fmt.Errorf("not enough data for header")
	}

	var result Header

	result.Data = append(result.Data, data...)

	if data[0] == 'M' && data[1] == 'M' {
		result.Endian = binary.BigEndian
	} else if data[0] == 'I' && data[1] == 'I' {
		result.Endian = binary.LittleEndian
	} else {
		result.Endian = nil
	}

	magic := result.Endian.Uint16(data[2:4])
	if magic == 43{
		return &result, 8, fmt.Errorf("parsing bigtiff file not yet implemented")
	}	
	if magic != 42 {
		return &result, 8, fmt.Errorf("not a tiff file, magic number was %v expected 42", magic)
	}
	result.FirstIFDOffset = int64(result.Endian.Uint32(data[4:8]))

	result.Start = 0
	result.End = 8

	return &result, int64(len(result.Data)), nil

}

func (h *Header) Contains(offset int64) bool {
	return h.Start <= offset && offset < h.End
}

func (h *Header) ContainsRegion(start int64, end int64) bool {
	return h.Start <= start && start < h.End && h.Start < end && end < h.End
}

func (h *Header) Find(offset int64) (parser.Region, error) {
	if offset < h.Start || offset >= h.End {
		return nil, fmt.Errorf("find offset %v outside of region %v to %v", offset, h.Start, h.End)
	}
	return h, nil
}

func (h *Header) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("header can not be split")
}

func (h *Header) Render() ([]payload.Section, error) {
	desc, err := payload.RenderTemplate(headerTemplate, h, template.FuncMap{})
	if err != nil {
		return nil, fmt.Errorf("could not render header description, %v", err)
	}

	var data bytes.Buffer

	payload.RenderBytesSpan(&data, h.Data[0:2], "header_endian")
	data.WriteRune(' ')
	payload.RenderBytesSpan(&data, h.Data[2:4], "header_magic")
	data.WriteRune(' ')
	payload.RenderBytesSpan(&data, h.Data[4:8], "header_offset")

	return []payload.Section{
		&payload.General{
			Start:   h.Start,
			End:     h.End - 1,
			Id:      "header",
			TheData: template.HTML(data.String()),
			Text:    template.HTML(desc),
		},
	}, nil
}

const headerTemplate = `The header shows this is a <span class="header_magic">tiff file</span> is in <span class="header_endian">{{.Endian}}</span> format {{if .BigTiff}} and is a big tiff {{end}} The first IDF starts at byte <span class="header_offset"><a href="#{{.FirstIFDOffset}}">{{.FirstIFDOffset}}</a></span>`
