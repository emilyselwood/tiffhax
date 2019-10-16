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
	if err := startRegion.Split(0, l, header); err != nil {
		return nil, fmt.Errorf("could not insert header %v", err)
	}

	return startRegion.Render()
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

