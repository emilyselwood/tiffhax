package parser

import (
	"fmt"
	"github.com/emilyselwood/tiffhax/payload"
	"strconv"
)

type Region interface {
	Split(start int64, end int64, newBit Region) error
	Render() ([]payload.Section, error)
	Find(offset int64) (Region, error)
	Contains(offset int64) bool
	ContainsRegion(start int64, end int64) bool
}

type Unknown struct {
	Start int64
	End   int64

	Children []Region
}

func (u *Unknown) Contains(offset int64) bool {
	return u.Start <= offset && offset < u.End
}

func (u *Unknown) ContainsRegion(start int64, end int64) bool {
	return u.Start <= start && start < u.End && u.Start < end && end <= u.End
}

func (u *Unknown) Find(offset int64) (Region, error) {
	if offset < u.Start || offset >= u.End {
		return nil, fmt.Errorf("find offset %v outside of region %v to %v", offset, u.Start, u.End)
	}

	if len(u.Children) > 0 {
		for _, c := range u.Children {
			if c.Contains(offset) {
				r, err := c.Find(offset)
				return r, err
			}
		}
		return nil, fmt.Errorf("find offset %v inside region %v to %v but not contained by children ... bork bork bork", offset, u.Start, u.End)
	} else {
		return u, nil
	}
}

func (u *Unknown) Split(start int64, end int64, newBit Region) error {
	if !u.ContainsRegion(start, end) {
		return fmt.Errorf("split region %v to %v outside of region %v to %v", start, end, u.Start, u.End)
	}
	if len(u.Children) != 0 {
		return fmt.Errorf("split region %v to %v is in %v to %v but it has children already", start, end, u.Start, u.End)
	}

	if start == u.Start && end == u.End {
		// Really we should replace this Unknown in its parent but this works for now.
		u.Children = []Region{
			newBit,
		}
	} else if start == u.Start {
		u.Children = []Region{
			newBit,
			&Unknown{
				Start: end,
				End:   u.End,
			},
		}
	} else if end == u.End {
		u.Children = []Region{
			&Unknown{
				Start: u.Start,
				End:   start,
			},
			newBit,
		}
	} else {
		u.Children = []Region{
			&Unknown{
				Start: u.Start,
				End:   start,
			},
			newBit,
			&Unknown{
				Start: end,
				End:   u.End,
			},
		}
	}

	return nil
}

func (u *Unknown) Render() ([]payload.Section, error) {
	if len(u.Children) > 0 {
		var result []payload.Section
		for _, c := range u.Children {
			part, err := c.Render()
			if err != nil {
				return nil, err
			}
			result = append(result, part...)
		}
		return result, nil
	} else {
		// TODO better rendering.
		return []payload.Section{&payload.General{Start: u.Start, End: u.End, Id: strconv.FormatInt(u.Start, 10)}}, nil
	}
}
