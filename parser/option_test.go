package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yoheimuta/go-protoparser/internal/lexer"
	"github.com/yoheimuta/go-protoparser/parser"
	"github.com/yoheimuta/go-protoparser/parser/meta"
)

func TestParser_ParseOption(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantOption *parser.Option
		wantErr    bool
	}{
		{
			name:    "parsing an empty",
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without option",
			input:   `java_package = "com.example.foo";`,
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without =",
			input:   `option java_package "com.example.foo";`,
			wantErr: true,
		},
		{
			name:    "parsing an invalid; without ;",
			input:   `option java_package = "com.example.foo"`,
			wantErr: true,
		},
		{
			name:  "parsing an excerpt from the official reference",
			input: `option java_package = "com.example.foo";`,
			wantOption: &parser.Option{
				OptionName: "java_package",
				Constant:   `"com.example.foo"`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:  "parsing another excerpt from the official reference",
			input: `option (my_option).a = true;`,
			wantOption: &parser.Option{
				OptionName: "(my_option).a",
				Constant:   `true`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
		{
			name:  "parsing fullIdent",
			input: `option java_package.baz.bar = "com.example.foo";`,
			wantOption: &parser.Option{
				OptionName: "java_package.baz.bar",
				Constant:   `"com.example.foo"`,
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 1,
						Line:   1,
						Column: 1,
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			p := parser.NewParser(lexer.NewLexer(strings.NewReader(test.input)))
			got, err := p.ParseOption()
			switch {
			case test.wantErr:
				if err == nil {
					t.Errorf("got err nil, but want err")
				}
				return
			case !test.wantErr && err != nil:
				t.Errorf("got err %v, but want nil", err)
				return
			}

			if !reflect.DeepEqual(got, test.wantOption) {
				t.Errorf("got %v, but want %v", got, test.wantOption)
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}

}
