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

type IFD struct {
	Start int64
	End   int64
	HeaderData  []byte
	FooterData  []byte
	Count uint16
	Children []*Field
	Next uint32
}

func ParseIFD(in io.Reader, start int64, order binary.ByteOrder) (*IFD, int64, []*Offset, error) {
	ifdHeader := make([]byte, 2)

	n, err := in.Read(ifdHeader)
	if err != nil {
		return nil, int64(n), nil, fmt.Errorf("could not read ifd header, %v", err)
	}

	var result IFD

	result.Start = start
	result.Count = order.Uint16(ifdHeader)
	result.End = start + 6 + (int64(result.Count) * 12)
	result.HeaderData = ifdHeader

	// Now read the fields for the IFD
	var offsets []*Offset
	for i := 0; i < int(result.Count); i++ {
		fieldStart := start + 2 + (int64(i) * 12)
		field, offset, err := ParseField(in, fieldStart, order)
		if err != nil {
			return nil, 0, nil, fmt.Errorf("could not parse field %v of ifd, %v", i, err)
		}
		result.Children = append(result.Children, field)
		if offset != nil {
			offsets = append(offsets, offset)
		}
	}

	nextIFD := make([]byte, 4)
	n, err = in.Read(nextIFD)
	if err != nil {
		return nil, int64(n), nil, fmt.Errorf("could not read ifd footer, %v", err)
	}

	result.Next = order.Uint32(nextIFD)
	result.FooterData = nextIFD
	return &result, result.End, offsets, nil
}


func (i *IFD) Contains(offset int64) bool {
	return i.Start <= offset && offset < i.End
}

func (i *IFD) ContainsRegion(start int64, end int64) bool {
	return i.Start <= start && start < i.End && i.Start < end && end < i.End
}

func (i *IFD) Find(offset int64) (parser.Region, error) {
	if offset < i.Start || offset >= i.End {
		return nil, fmt.Errorf("find offset %v outside of ifd region %v to %v", offset, i.Start, i.End)
	}
	if offset >= i.Start + 2 {
		if len(i.Children) > 0 {
			for _, c := range i.Children {
				if c.Contains(offset) {
					r, err := c.Find(offset)
					return r, err
				}
			}
			return nil, fmt.Errorf("find offset %v inside region %v to %v but not contained by children ... bork bork bork", offset, i.Start, i.End)
		} else {
			return nil, fmt.Errorf("find offset past ifd header but ifd does not have any children ... bork bork bork")
		}
	}

	return i, nil
}

func (i *IFD) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("IFD can not be split")
}

func (i *IFD) Render() ([]payload.Section, error) {
	var result []payload.Section

	// ifd header
	header, err := i.renderHeader()
	if err != nil {
		return result, fmt.Errorf("could not render ifd header, %v", err)
	}
	result = append(result, header)

	// each field
	for _, f := range i.Children {
		childSections, err := f.Render()
		if err != nil {
			return result, fmt.Errorf("could not render ifd field, %v", err)
		}
		result = append(result, childSections...)
	}

	// ifd footer
	footer, err := i.renderFooter()
	if err != nil {
		return result, fmt.Errorf("could not render ifd footer, %v", err)
	}
	result = append(result, footer)

	return result, nil
}

func (i *IFD) renderHeader() (payload.Section, error) {

	desc, err := payload.RenderTemplate(ifdHeaderTemplate, i, template.FuncMap{})
	if err != nil {
		return nil, fmt.Errorf("could not render ifd description, %v", err)
	}

	var data bytes.Buffer
	payload.RenderBytesSpan(&data, i.HeaderData, "ifd_header")

	return &payload.General{
		Start:   i.Start,
		End:     i.Start + 1,
		Id:      "ifd",
		TheData: template.HTML(data.String()),
		Text:    template.HTML(desc),
	}, nil
}
const ifdHeaderTemplate = `The start of an IFD (Image File Directory) that contains <span class="ifd_header">{{.Count}}</span> fields`


func (i *IFD) renderFooter() (payload.Section, error) {
	var desc string
	var data bytes.Buffer

	if i.Next == 0 {
		desc = "This is the last IFD in the chain. If this <span class=\"ifd_footer\">0</span> was a number it would point to the next ifd"
	} else {
		desc = fmt.Sprintf("The next ifd can be found at offset <span class=\"ifd_footer\">%v</span>", i.Next)
	}

	payload.RenderBytesSpan(&data, i.FooterData, "ifd_footer")

	return &payload.General{
		Start:   i.End - 5,
		End:     i.End - 1,
		Id:      "ifd",
		TheData: template.HTML(data.String()),
		Text:    template.HTML(desc),
	}, nil
}