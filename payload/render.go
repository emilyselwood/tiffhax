package payload

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strconv"
	"strings"
)

func RenderTemplate(t string, data interface{}, funcMap template.FuncMap) (string, error) {
	tmpl, err := template.New("template").Funcs(funcMap).Parse(t)
	if err != nil {
		return "", fmt.Errorf("could not parse template, %v", err)
	}

	var descBuffer bytes.Buffer
	if err := tmpl.Execute(&descBuffer, data); err != nil {
		return "", fmt.Errorf("could not render template, %v", err)
	}

	return descBuffer.String(), nil
}

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
			i := 1
			for _, b := range in[1:] {
				buffer.WriteString(" ")

				buffer.WriteString(RenderByte(b))
				i++
				if i == 16 {
					buffer.WriteString("<br />")
					i = 0
				}
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