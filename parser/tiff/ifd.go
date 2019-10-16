package tiff

import (
	"encoding/binary"
	"fmt"
	"github.com/emilyselwood/tiffhax/parser"
	"github.com/emilyselwood/tiffhax/payload"
	"io"
)

type IFD struct {
	Start int64
	End   int64
	Data  []byte
	Count uint16
	Children []*Field
	Next uint32
}

func ParseIFD(in io.Reader, start int64, order binary.ByteOrder) (*IFD, int64, error) {
	ifdHeader := make([]byte, 2)

	n, err := in.Read(ifdHeader)
	if err != nil {
		return nil, int64(n), fmt.Errorf("could not read ifd header, %v", err)
	}

	var result IFD

	result.Start = start
	result.Count = order.Uint16(ifdHeader)
	result.End = start + 6 + (count * 12)

	// Now read the fields for the IFD

	for i := 0; i < int(result.Count); i++ {
		fieldStart := start + 2 + (int64(i) * 12)
		field, offset, err := ParseField(in, fieldStart, order)

	}


	nextIFD := make([]byte, 4)
	n, err = in.Read(ifdHeader)
	if err != nil {
		return nil, int64(n), fmt.Errorf("could not read ifd footer, %v", err)
	}

	result.Next = order.Uint32(nextIFD)
	return &result, result.End, nil
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
	return i, nil
}

func (i *IFD) Split(start int64, end int64, newBit parser.Region) error {
	return fmt.Errorf("IFD can not be split")
}

func (i *IFD) Render() ([]payload.Section, error) {

}