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

type Data struct {
	IFD   *IFD
	Start int64
	End   int64
	DType uint16
	Count int64
	I     int
}

func (d *Data) Parse(in io.ReadSeeker, order binary.ByteOrder) error {
	byteCounts, err := d.fetchFieldValue(279, in, order) // strip byte counts
	if err != nil {
		// try for the other field
		byteCounts, err = d.fetchFieldValue(325, in, order)
		if err != nil {
			return fmt.Errorf("could not find rows per strip field, %v", err)
		}
	}

	d.End = d.Start + byteCounts

	return nil
}

func (d *Data) Contains(offset int64) bool {
	return d.Start <= offset && offset < d.End
}

func (d *Data) ContainsRegion(start int64, end int64) bool {
	return d.Start <= start && start < d.End && d.Start < end && end < d.End
}

func (d *Data) Find(offset int64) (parser.Region, error) {
	if offset < d.Start || offset >= d.End {
		return nil, fmt.Errorf("find offset %v outside of offset region %v to %v", offset, d.Start, d.End)
	}
	return d, nil
}

func (d *Data) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("offset can not be split")
}

func (d *Data) Render() ([]payload.Section, error) {
	return []payload.Section{
		&payload.General{
			Start:   d.Start,
			End:     d.End - 1,
			Id:      "data",
			TheData: template.HTML("data hidden for size"),
			Text:    template.HTML("A block of image data"),
		},
	}, nil
}

func (d *Data) fetchFieldValue(id uint16, in io.ReadSeeker, order binary.ByteOrder) (int64, error) {
	field, err := d.IFD.FindField(id)
	if err != nil {
		return 0, fmt.Errorf("could not find field %v, %v", id, err)
	}

	if field.IsOffset {
		// now we need to go and find the lookup value for d.I
		// work out where we need to jump to

		pos := int64(field.Value) + int64(constants.DataTypeSize[field.DType] * uint32(d.I))

		_, err := in.Seek(pos, io.SeekStart)
		if err != nil {
			return 0, fmt.Errorf("could not seek to find offset value, %v", err)
		}

		buf := make([]byte, constants.DataTypeSize[field.DType])
		n, err := in.Read(buf)
		if err != nil {
			return 0, fmt.Errorf("could not read offset value, %v", err)
		}
		if n != int(constants.DataTypeSize[field.DType]) {
			return 0, fmt.Errorf("reading offset got the wrong number of bytes, expected %v got %v",constants.DataTypeSize[field.DType], n)
		}
		return int64(ReadBuffer(buf, order)), nil


	}

	return int64(field.Value), nil
}


func ReadBuffer(buf []byte, order binary.ByteOrder) uint32 {
	len := len(buf)

	if len == 2 {
		return uint32(order.Uint16(buf))
	}
	if len == 4 {
		return order.Uint32(buf)
	}
	return 0
}