package parser

import (
	"bufio"
	"fmt"
	"io"
)

type RespWriter struct {
	writer *bufio.Writer
}

func NewWriter(w io.Writer) *RespWriter {
	return &RespWriter{bufio.NewWriter(w)}
}

func (w *RespWriter) Write(v Value) error {
	b, err := v.Marshal()
	if err != nil {
		return err
	}
	_, err = w.writer.Write(b)
	if err != nil {
		return err
	}

	return w.writer.Flush()
}

func (v Value) Marshal() ([]byte, error) {
	switch v.RType {
	case INTEGER:
		return v.marshalInteger()
	case STRING:
		return v.marshalString()
	case ERROR:
		return v.marshalError()
	case BULK:
		return v.marshalBulk()
	case ARRAY:
		return v.marshalArray()
	case NULL:
		return v.marshalNull()
	default:
		return []byte{}, fmt.Errorf("invalid type: %q", v.RType)
	}
}

func (v Value) marshalString() ([]byte, error) {
	return fmt.Appendf([]byte{}, "+%s\r\n", v.Str), nil
}

func (v Value) marshalError() ([]byte, error) {
	return fmt.Appendf([]byte{}, "-%s\r\n", v.Str), nil
}

func (v Value) marshalInteger() ([]byte, error) {
	return fmt.Appendf([]byte{}, ":%d\r\n", v.Num), nil
}

func (v Value) marshalBulk() ([]byte, error) {
	return fmt.Appendf([]byte{}, "$%d\r\n%s\r\n", len(v.Bulk), v.Bulk), nil
}

func (v Value) marshalNull() ([]byte, error) {
	return []byte("$-1\r\n"), nil
}

func (v Value) marshalArray() ([]byte, error) {
	buf := fmt.Appendf([]byte{}, "*%d\r\n", len(v.Array))

	for i := range len(v.Array) {
		gotBuf, err := v.Array[i].Marshal()
		if err != nil {
			return []byte{}, err
		}
		buf = append(buf, gotBuf...)
	}

	return buf, nil
}
