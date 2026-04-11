package parser

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
	NULL    = '_'
)

type Value struct {
	RType rune
	Str   string
	Bulk  string
	Num   int
	Array []Value
}
