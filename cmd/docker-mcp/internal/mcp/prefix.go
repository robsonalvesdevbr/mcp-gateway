package mcp

import (
	"bytes"
	"io"
)

type Prefixer struct {
	prefix          string
	writer          io.Writer
	trailingNewline bool
	buf             bytes.Buffer
}

func newPrefixer(writer io.Writer, prefix string) *Prefixer {
	return &Prefixer{
		prefix:          prefix,
		writer:          writer,
		trailingNewline: true,
	}
}

func (pf *Prefixer) Write(payload []byte) (int, error) {
	pf.buf.Reset()

	for _, b := range payload {
		if pf.trailingNewline {
			pf.buf.WriteString(pf.prefix)
			pf.trailingNewline = false
		}

		pf.buf.WriteByte(b)

		if b == '\n' {
			pf.trailingNewline = true
		}
	}

	n, err := pf.writer.Write(pf.buf.Bytes())
	if err != nil {
		if n > len(payload) {
			n = len(payload)
		}
		return n, err
	}

	return len(payload), nil
}
