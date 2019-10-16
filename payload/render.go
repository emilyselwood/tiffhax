package payload

import (
	"bytes"
	"io"
	"strconv"
	"strings"
)

/*
RenderBytesSpan writes a byte array formatted as hex with a surrounding span.
This is useful for highlighting sections of the data.
 */
func RenderBytesSpan(target io.StringWriter, in []byte, class string) io.StringWriter {
	_, _ = target.WriteString("<span class=\"")
	_, _ = target.WriteString(class)
	_, _ = target.WriteString("\">")
	_, _ = target.WriteString(RenderBytes(in))
	_, _ = target.WriteString("</span>")

	return target
}

/*
RenderBytes displays an array of bytes in hex with spaces between each byte
 */
func RenderBytes(in []byte) string {
	var buffer bytes.Buffer
	if len(in) > 0 {
		buffer.WriteString(RenderByte(in[0]))
		if len(in) > 1 {
			for _, b := range in[1:] {
				buffer.WriteString(" ")

				buffer.WriteString(RenderByte(b))
			}
		}
	}

	return strings.ToUpper(buffer.String())
}

func RenderByte(in byte) string {
	result := strconv.FormatInt(int64(in), 16)
	if len(result) < 2 {
		result = "0" + result
	}
	return result
}