package parser

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	RType string
	Str   string
	Bulk  string
	Num   int
	Array []Value
}
