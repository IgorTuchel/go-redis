package parser

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Resp struct {
	reader *bufio.Reader
}

func NewResp(r io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(r)}
}

func (r *Resp) Read() (Value, error) {
	rType, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch rType {
	case INTEGER:
		return r.parseInt()
	case STRING:
		return r.parseString()
	case ERROR:
		return r.parseError()
	case BULK:
		return r.parseBulk()
	case ARRAY:
		return r.parseArray()
	default:
		return Value{}, fmt.Errorf("unkown type: %q", rType)
	}
}

func (r *Resp) readLine() ([]byte, error) {
	line, err := r.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line[:len(line)-2], nil
}

func (r *Resp) readInteger() (int, error) {
	line, err := r.readLine()
	if err != nil {
		return 0, err
	}
	intLine, err := strconv.Atoi(string(line))
	if err != nil {
		return 0, err
	}
	return intLine, nil
}

func (r *Resp) parseInt() (Value, error) {
	intLine, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}
	v := Value{RType: INTEGER, Num: intLine}

	return v, nil
}

func (r *Resp) parseString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	v := Value{RType: STRING, Str: string(line)}

	return v, nil
}

func (r *Resp) parseError() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	v := Value{RType: ERROR, Str: string(line)}

	return v, nil
}

func (r *Resp) parseBulk() (Value, error) {
	bulkLen, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}

	if bulkLen == -1 {
		return Value{RType: NULL}, nil
	}

	buf := make([]byte, bulkLen)

	_, err = io.ReadFull(r.reader, buf)
	if err != nil {
		return Value{}, err
	}

	_, err = r.readLine() // Consume the trailing CRLF
	if err != nil {
		return Value{}, err
	}

	v := Value{RType: BULK, Bulk: string(buf)}

	return v, nil
}

func (r *Resp) parseArray() (Value, error) {
	lenArray, err := r.readInteger()
	if err != nil {
		return Value{}, err
	}

	if lenArray == -1 {
		return Value{RType: NULL}, nil
	}

	v := Value{RType: ARRAY}

	for range lenArray {
		value, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		v.Array = append(v.Array, value)
	}

	return v, nil
}
