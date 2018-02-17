package chm

import (
	"bytes"
	"fmt"
	"log"
)

type Buffer struct {
	buffer bytes.Buffer
	indent string
}

func (b *Buffer) Write(s string) {
	_, err := b.buffer.WriteString(s)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Buffer) Line(p ...interface{}) {
	if len(p) == 0 {
		b.Write("\r\n")
	} else {
		b.Write(b.indent)
		if len(p) == 1 {
			b.Write(p[0].(string))
		} else {
			b.Write(fmt.Sprintf(p[0].(string), p[1:]...))
		}
		b.Write("\r\n")
	}
}

func (b *Buffer) Indent(p ...interface{}) {
	if len(p) > 0 {
		b.Line(p...)
	}
	b.indent += "\t"
}

func (b *Buffer) Unindent(p ...interface{}) {
	b.indent = b.indent[:len(p)-1]
	if len(p) > 0 {
		b.Line(p...)
	}
}

func (b *Buffer) Bytes() []byte {
	return b.buffer.Bytes()
}
