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
		return returnError(&startRegion, fmt.Errorf("could not parse header, %v", err))
	}
	if err := insert(&startRegion, header, 0, l); err != nil {
		return returnError(&startRegion, fmt.Errorf("could not insert header %v", err))
	}

	var offsets []*Offset
	var data []*Data
	// start with the first IFD (there must be at least one)
	ifd, offset, d, err := readIFD(in, header, header.FirstIFDOffset)
	if err != nil {
		return returnError(&startRegion, fmt.Errorf("could not read first ifd, %v", err))
	}

	if err := insert(&startRegion, ifd, ifd.Start, ifd.End); err != nil {
		return returnError(&startRegion, fmt.Errorf("could not insert ifd, %v", err))
	}
	offsets = append(offsets, offset...)
	data = append(data, d...)

	for ifd.Next != 0 {
		ifd, offset, d, err = readIFD(in, header, int64(ifd.Next))
		if err != nil {
			return returnError(&startRegion, fmt.Errorf("could not read ifd, %v", err))
		}

		if err := insert(&startRegion, ifd, ifd.Start, ifd.End); err != nil {
			return returnError(&startRegion, fmt.Errorf("could not insert ifd, %v", err))
		}
		offsets = append(offsets, offset...)
		data = append(data, d...)
	}

	// now handle the offsets
	// because an offset can point to a list of offsets we need to keep handling them till we are done.
	for _, o := range offsets {
		d, err := o.Parse(in, header.Endian)
		if err != nil {
			return returnError(&startRegion, fmt.Errorf("could not parse offset, %v", err))
		}
		if err := insert(&startRegion, o, o.Start, o.End); err != nil {
			return returnError(&startRegion, fmt.Errorf("could not insert offset result, %v", err))
		}
		data = append(data, d...)
	}

	// Finally we need to handle the data sections.
	// Going to need to be able to work out:
	//  a: where the strips start
	//  b: how big each strip is.

	for _, d := range data {
		err := d.Parse(in, header.Endian)
		if err != nil {
			return returnError(&startRegion, fmt.Errorf("could not parse data information, %v", err))
		}
		if err := insert(&startRegion, d, d.Start, d.End); err != nil {
			return returnError(&startRegion, fmt.Errorf("could not insert data result, %v", err))
		}
	}


	return startRegion.Render()
}

func returnError(startRegion parser.Region, inErr error) ([]payload.Section, error) {
	res, err := startRegion.Render();
	if err != nil {
		return res, fmt.Errorf("%v, additionaly while rendering the output the following happend: %v", inErr, err)
	}
	return res, inErr
}

func readIFD(in io.ReadSeeker, header *Header, offset int64) (*IFD, []*Offset, []*Data, error) {
	_, err := in.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not seek to IFD, %v", err)
	}

	ifd, _, offsets, d, err := ParseIFD(in, offset, header.Endian)
	if err != nil {
		return  nil, nil, nil, fmt.Errorf("could not parse IFD, %v", err)
	}

	return ifd, offsets, d, nil
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

