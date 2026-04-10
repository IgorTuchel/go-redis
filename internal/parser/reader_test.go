package parser

import (
	"bytes"
	"testing"
)

func TestRead(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{
			name:  "simple string",
			input: "+OK\r\n",
			want:  Value{RType: "STRING", Str: "OK"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewResp(bytes.NewBufferString(tc.input))
			got, err := r.Read()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.RType != tc.want.RType {
				t.Errorf("RType got %q, want %q", got.RType, tc.want.RType)
			}
			if got.Str != tc.want.Str {
				t.Errorf("Str: got %q, want %q", got.Str, tc.want.Str)
			}
		})
	}

}
