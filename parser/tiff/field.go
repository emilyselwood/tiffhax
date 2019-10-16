package tiff

import (
	"encoding/binary"
	"io"
)

type Field struct {
	Start int64
	End   int64
	Data  []byte
	ID    int32
	Count int32
	DType int32
	Value interface{}
}

func ParseField(in io.Reader, start int64, order binary.ByteOrder) (*Field, []Offset, error) {

}