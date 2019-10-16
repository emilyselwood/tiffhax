package tiff

import (
	"fmt"
	"github.com/emilyselwood/tiffhax/parser"
	"github.com/emilyselwood/tiffhax/payload"
	"io"
)

func ParseFile(in io.ReadSeeker) ([]payload.Section, error) {

	start, end, err := findExtents(in)
	if err != nil {
		return nil, fmt.Errorf("could not find reader extents: %v", err)
	}

	startRegion := parser.Unknown{
		Start:    start,
		End:      end,
		Children: []parser.Region{},
	}

	// start by parsing the header
	header, l, err := ParseHeader(in)
	if err != nil {
		return nil, err
	}
	if err := insert(&startRegion, header, 0, l); err != nil {
		return nil, fmt.Errorf("could not insert header %v", err)
	}

	var offsets []*Offset
	// start with the first IFD
	ifd, offset, err := readIFD(in, header, header.FirstIFDOffset)
	if err != nil {
		return nil, fmt.Errorf("could not read first ifd, %v", err)
	}

	if err := insert(&startRegion, ifd, ifd.Start, ifd.End); err != nil {
		return nil, fmt.Errorf("could not insert ifd, %v", err)
	}
	offsets = append(offsets, offset...)

	for ifd.Next != 0 {
		ifd, offset, err := readIFD(in, header, int64(ifd.Next))
		if err != nil {
			return nil, fmt.Errorf("could not read ifd, %v", err)
		}

		if err := insert(&startRegion, ifd, ifd.Start, ifd.End); err != nil {
			return nil, fmt.Errorf("could not insert ifd, %v", err)
		}
		offsets = append(offsets, offset...)
	}

	// now handle the offsets

	return startRegion.Render()
}

func readIFD(in io.ReadSeeker, header *Header, offset int64) (*IFD, []*Offset, error) {
	_, err := in.Seek(header.FirstIFDOffset, io.SeekStart)
	if err != nil {
		return nil, nil, fmt.Errorf("could not seek to IFD, %v", err)
	}

	ifd, _, offsets, err := ParseIFD(in, header.FirstIFDOffset, header.Endian)
	if err != nil {
		return  nil, nil, fmt.Errorf("could not parse IFD, %v", err)
	}

	return ifd, offsets, nil
}

func insert(top parser.Region, newBit parser.Region, start int64, end int64) error {
	target, err := top.Find(start)
	if err != nil {
		return fmt.Errorf("could not insert new region, %v", err)
	}
	if err := target.Split(start, end, newBit); err != nil {
		return fmt.Errorf("could not split, %v", err)
	}

	return nil
}

func findExtents(in io.ReadSeeker) (int64, int64, error) {
	end, err := in.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, fmt.Errorf("could not seek to end of file %v", err)
	}
	start, err := in.Seek(0, io.SeekStart)
	if err != nil {
		return 0, 0, fmt.Errorf("could not seek to start of file %v", err)
	}

	return start, end, nil
}

