package parser

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReadString(t *testing.T) {
	r := NewResp(bytes.NewBufferString("+OK\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, STRING, got.RType)
	assert.Equal(t, "OK", got.Str)

	r = NewResp(bytes.NewBufferString("+Hello\r\n"))
	got, err = r.Read()
	require.NoError(t, err)
	assert.Equal(t, STRING, got.RType)
	assert.Equal(t, "Hello", got.Str)
}

func TestReadError(t *testing.T) {
	r := NewResp(bytes.NewBufferString("-ERR unknown command\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, ERROR, got.RType)
	assert.Equal(t, "ERR unknown command", got.Str)
}

func TestReadInteger(t *testing.T) {
	r := NewResp(bytes.NewBufferString(":42\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, INTEGER, got.RType)
	assert.Equal(t, 42, got.Num)

	// negative integer
	r = NewResp(bytes.NewBufferString(":-1\r\n"))
	got, err = r.Read()
	require.NoError(t, err)
	assert.Equal(t, INTEGER, got.RType)
	assert.Equal(t, -1, got.Num)
}

func TestReadBulk(t *testing.T) {
	r := NewResp(bytes.NewBufferString("$5\r\nhello\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, BULK, got.RType)
	assert.Equal(t, "hello", got.Bulk)

	// empty bulk string
	r = NewResp(bytes.NewBufferString("$0\r\n\r\n"))
	got, err = r.Read()
	require.NoError(t, err)
	assert.Equal(t, BULK, got.RType)
	assert.Equal(t, "", got.Bulk)
}

func TestReadNullBulk(t *testing.T) {
	r := NewResp(bytes.NewBufferString("$-1\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, NULL, got.RType)
}

func TestReadArray(t *testing.T) {
	// SET foo bar
	r := NewResp(bytes.NewBufferString("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"))
	got, err := r.Read()
	require.NoError(t, err)
	assert.Equal(t, ARRAY, got.RType)
	require.Len(t, got.Array, 3)
	assert.Equal(t, "SET", got.Array[0].Bulk)
	assert.Equal(t, "foo", got.Array[1].Bulk)
	assert.Equal(t, "bar", got.Array[2].Bulk)

	// null array
	r = NewResp(bytes.NewBufferString("*-1\r\n"))
	got, err = r.Read()
	require.NoError(t, err)
	assert.Equal(t, NULL, got.RType)
}
