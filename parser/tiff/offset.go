package tiff

/*
Offset holds a link from somewhere in a file to somewhere else.
*/
type Offset struct {
	From  int64
	To    int64
	DType int
}
