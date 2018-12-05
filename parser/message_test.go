package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/yoheimuta/go-protoparser/internal/lexer"
	"github.com/yoheimuta/go-protoparser/internal/util_test"
	"github.com/yoheimuta/go-protoparser/parser"
	"github.com/yoheimuta/go-protoparser/parser/meta"
)

func TestParser_ParseMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMessage *parser.Message
		wantErr     bool
	}{
		{
			name:    "parsing an empty",
			wantErr: true,
		},
		{
			name: "parsing an excerpt from the official reference",
			input: `
message Outer {
  option (my_option).a = true;
  message Inner {
    int64 ival = 1;
  }
  map<int32, string> my_map = 2;
}
`,
			wantMessage: &parser.Message{
				MessageName: "Outer",
				MessageBody: []interface{}{
					&parser.Option{
						OptionName: "(my_option).a",
						Constant:   "true",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 20,
								Line:   3,
								Column: 3,
							},
						},
					},
					&parser.Message{
						MessageName: "Inner",
						MessageBody: []interface{}{
							&parser.Field{
								Type:        "int64",
								FieldName:   "ival",
								FieldNumber: "1",
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 71,
										Line:   5,
										Column: 5,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 51,
								Line:   4,
								Column: 3,
							},
						},
					},
					&parser.MapField{
						KeyType:     "int32",
						Type:        "string",
						MapName:     "my_map",
						FieldNumber: "2",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 93,
								Line:   7,
								Column: 3,
							},
						},
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 2,
						Line:   2,
						Column: 1,
					},
				},
			},
		},
		{
			name: "parsing another excerpt from the official reference",
			input: `
message outer {
  option (my_option).a = true;
  message inner {
    int64 ival = 1;
  }
  repeated inner inner_message = 2;
  EnumAllowingAlias enum_field =3;
  map<int32, string> my_map = 4;
}
`,
			wantMessage: &parser.Message{
				MessageName: "outer",
				MessageBody: []interface{}{
					&parser.Option{
						OptionName: "(my_option).a",
						Constant:   "true",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 20,
								Line:   3,
								Column: 3,
							},
						},
					},
					&parser.Message{
						MessageName: "inner",
						MessageBody: []interface{}{
							&parser.Field{
								Type:        "int64",
								FieldName:   "ival",
								FieldNumber: "1",
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 71,
										Line:   5,
										Column: 5,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 51,
								Line:   4,
								Column: 3,
							},
						},
					},
					&parser.Field{
						IsRepeated:  true,
						Type:        "inner",
						FieldName:   "inner_message",
						FieldNumber: "2",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 93,
								Line:   7,
								Column: 3,
							},
						},
					},
					&parser.Field{
						Type:        "EnumAllowingAlias",
						FieldName:   "enum_field",
						FieldNumber: "3",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 129,
								Line:   8,
								Column: 3,
							},
						},
					},
					&parser.MapField{
						KeyType:     "int32",
						Type:        "string",
						MapName:     "my_map",
						FieldNumber: "4",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 164,
								Line:   9,
								Column: 3,
							},
						},
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 2,
						Line:   2,
						Column: 1,
					},
				},
			},
		},
		{
			name: "parsing an empty MessageBody",
			input: `
message Outer {
}
`,
			wantMessage: &parser.Message{
				MessageName: "Outer",
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 2,
						Line:   2,
						Column: 1,
					},
				},
			},
		},
		{
			name: "parsing comments",
			input: `
message outer {
  // option
  option (my_option).a = true;
  // message
  message inner {   // Level 2
    int64 ival = 1;
  }
  // field
  repeated inner inner_message = 2;
  // enum
  enum EnumAllowingAlias {
    option allow_alias = true;
  }
  EnumAllowingAlias enum_field =3;
  // map
  map<int32, string> my_map = 4;
  // oneof
  oneof foo {
    string name = 5;
    SubMessage sub_message = 6;
  }
  // reserved
  reserved "bar";
}
`,
			wantMessage: &parser.Message{
				MessageName: "outer",
				MessageBody: []interface{}{
					&parser.Option{
						OptionName: "(my_option).a",
						Constant:   "true",
						Comments: []*parser.Comment{
							{
								Raw: `// option`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 20,
										Line:   3,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 32,
								Line:   4,
								Column: 3,
							},
						},
					},
					&parser.Message{
						MessageName: "inner",
						MessageBody: []interface{}{
							&parser.Field{
								Type:        "int64",
								FieldName:   "ival",
								FieldNumber: "1",
								Comments: []*parser.Comment{
									{
										Raw: `// Level 2`,
										Meta: meta.Meta{
											Pos: meta.Position{
												Offset: 94,
												Line:   6,
												Column: 21,
											},
										},
									},
								},
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 109,
										Line:   7,
										Column: 5,
									},
								},
							},
						},
						Comments: []*parser.Comment{
							{
								Raw: `// message`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 63,
										Line:   5,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 76,
								Line:   6,
								Column: 3,
							},
						},
					},
					&parser.Field{
						IsRepeated:  true,
						Type:        "inner",
						FieldName:   "inner_message",
						FieldNumber: "2",
						Comments: []*parser.Comment{
							{
								Raw: `// field`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 131,
										Line:   9,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 142,
								Line:   10,
								Column: 3,
							},
						},
					},
					&parser.Enum{
						EnumName: "EnumAllowingAlias",
						EnumBody: []interface{}{
							&parser.Option{
								OptionName: "allow_alias",
								Constant:   "true",
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 217,
										Line:   13,
										Column: 5,
									},
								},
							},
						},
						Comments: []*parser.Comment{
							{
								Raw: `// enum`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 178,
										Line:   11,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 188,
								Line:   12,
								Column: 3,
							},
						},
					},
					&parser.Field{
						Type:        "EnumAllowingAlias",
						FieldName:   "enum_field",
						FieldNumber: "3",
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 250,
								Line:   15,
								Column: 3,
							},
						},
					},
					&parser.MapField{
						KeyType:     "int32",
						Type:        "string",
						MapName:     "my_map",
						FieldNumber: "4",
						Comments: []*parser.Comment{
							{
								Raw: `// map`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 285,
										Line:   16,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 294,
								Line:   17,
								Column: 3,
							},
						},
					},
					&parser.Oneof{
						OneofFields: []*parser.OneofField{
							{
								Type:        "string",
								FieldName:   "name",
								FieldNumber: "5",
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 354,
										Line:   20,
										Column: 5,
									},
								},
							},
							{
								Type:        "SubMessage",
								FieldName:   "sub_message",
								FieldNumber: "6",
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 375,
										Line:   21,
										Column: 5,
									},
								},
							},
						},
						OneofName: "foo",
						Comments: []*parser.Comment{
							{
								Raw: `// oneof`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 327,
										Line:   18,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 338,
								Line:   19,
								Column: 3,
							},
						},
					},
					&parser.Reserved{
						FieldNames: []string{
							`"bar"`,
						},
						Comments: []*parser.Comment{
							{
								Raw: `// reserved`,
								Meta: meta.Meta{
									Pos: meta.Position{
										Offset: 409,
										Line:   23,
										Column: 3,
									},
								},
							},
						},
						Meta: meta.Meta{
							Pos: meta.Position{
								Offset: 423,
								Line:   24,
								Column: 3,
							},
						},
					},
				},
				Meta: meta.Meta{
					Pos: meta.Position{
						Offset: 2,
						Line:   2,
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
			got, err := p.ParseMessage()
			switch {
			case test.wantErr:
				if err == nil {
					t.Errorf("got err nil, but want err, parsed=%v", got)
				}
				return
			case !test.wantErr && err != nil:
				t.Errorf("got err %v, but want nil", err)
				return
			}

			if !reflect.DeepEqual(got, test.wantMessage) {
				t.Errorf("got %v, but want %v", util_test.PrettyFormat(got), util_test.PrettyFormat(test.wantMessage))
			}

			if !p.IsEOF() {
				t.Errorf("got not eof, but want eof")
			}
		})
	}

}
