package payload

import (
	"fmt"
	"html/template"
)

type Payload struct {
	Title    string
	FileName string
	Sections []Section
}

type Section interface {
	ID() string
	Offset() string
	Data() template.HTML
	Description() template.HTML
}

/*
General represents a normal section of the file
*/
type General struct {
	Start int64
	End int64
	Id string
	TheData template.HTML
	Text template.HTML
}

func (g *General) ID() string {
	return g.Id
}

func (g *General) Offset() string {
	return fmt.Sprintf("%v .. %v", g.Start, g.End)
}

func (g *General) Data() template.HTML {
	return g.TheData
}

func (g *General) Description() template.HTML {
	return g.Text
}
