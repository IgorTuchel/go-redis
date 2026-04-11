package parser

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWriteString(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: STRING, Str: "OK"})
	require.NoError(t, err)
	assert.Equal(t, "+OK\r\n", buf.String())

	buf.Reset()
	err = w.Write(Value{RType: STRING, Str: "Hello"})
	require.NoError(t, err)
	assert.Equal(t, "+Hello\r\n", buf.String())
}

func TestWriteError(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: ERROR, Str: "ERR unknown command"})
	require.NoError(t, err)
	assert.Equal(t, "-ERR unknown command\r\n", buf.String())
}

func TestWriteInteger(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: INTEGER, Num: 42})
	require.NoError(t, err)
	assert.Equal(t, ":42\r\n", buf.String())

	buf.Reset()
	err = w.Write(Value{RType: INTEGER, Num: -1})
	require.NoError(t, err)
	assert.Equal(t, ":-1\r\n", buf.String())
}

func TestWriteBulk(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: BULK, Bulk: "hello"})
	require.NoError(t, err)
	assert.Equal(t, "$5\r\nhello\r\n", buf.String())

	// empty bulk string
	buf.Reset()
	err = w.Write(Value{RType: BULK, Bulk: ""})
	require.NoError(t, err)
	assert.Equal(t, "$0\r\n\r\n", buf.String())
}

func TestWriteNull(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: NULL})
	require.NoError(t, err)
	assert.Equal(t, "$-1\r\n", buf.String())
}

func TestWriteArray(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriter(buf)
	err := w.Write(Value{RType: ARRAY, Array: []Value{
		{RType: BULK, Bulk: "SET"},
		{RType: BULK, Bulk: "foo"},
		{RType: BULK, Bulk: "bar"},
	}})
	require.NoError(t, err)
	assert.Equal(t, "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n", buf.String())
}
