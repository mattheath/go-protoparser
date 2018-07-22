package protoparser

import (
	"fmt"
	"strings"
	"text/scanner"
)

type lexer struct {
	scan  scanner.Scanner
	token rune
}

func (lex *lexer) next()        { lex.token = lex.scan.Scan() }
func (lex *lexer) text() string { return lex.scan.TokenText() }

// Type はフィールドの型を表す。
type Type struct {
	Name       string
	IsRepeated bool
}

// Field は型のフィールドを表す。
type Field struct {
	Comments []string
	Type     *Type
	Name     string
}

// Oneof は oneof 型を表す。
type Oneof struct {
	Comments []string
	Name     string
	Fields   []*Field
}

// EnumField は Enum の値を表す。
type EnumField struct {
	Comments []string
	Name     string
}

// Enum は Enum 型を表す。
type Enum struct {
	Comments   []string
	Name       string
	EnumFields []*EnumField
}

// Message は独自に定義した型情報を表す。
type Message struct {
	Comments []string
	Name     string
	Fields   []*Field
	Nests    []*Message
	Enums    []*Enum
	Oneofs   []*Oneof
}

// RPC は関数を表す。
type RPC struct {
	Comments []string
	Name     string
	Argument *Type
	Return   *Type
}

// Service は複数の RPC を定義するサービスを表す。
type Service struct {
	Comments []string
	Name     string
	RPCs     []*RPC
}

// ProtocolBuffer は Protocol Buffers ファイルをパースした結果を表す。
type ProtocolBuffer struct {
	Package  string
	Service  *Service
	Messages []*Message
}

// Parse は Protocol Bufffers ファイルをパースする。
func Parse(input string) (*ProtocolBuffer, error) {
	lex := new(lexer)
	lex.scan.Init(strings.NewReader(input))
	lex.scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanFloats | scanner.ScanComments
	lex.next()
	return parse(lex)
}

// comment\nmessage...
func parse(lex *lexer) (*ProtocolBuffer, error) {
	var pkg string
	service := &Service{}
	var messages []*Message
	for lex.token != scanner.EOF {
		comments := parseComments(lex)

		switch lex.text() {
		case "package":
			p, err := parsePackage(lex)
			if err != nil {
				return nil, err
			}
			pkg = p
		case "service":
			s, err := parseService(lex)
			if err != nil {
				return nil, err
			}
			s.Comments = append(s.Comments, comments...)
			service = s
		case "message":
			message, err := parseMessage(lex)
			if err != nil {
				return nil, err
			}
			message.Comments = append(message.Comments, comments...)
			messages = append(messages, message)
		default:
			lex.next()
			continue
		}
	}
	return &ProtocolBuffer{
		Package:  pkg,
		Service:  service,
		Messages: messages,
	}, nil
}

// 'package' var';'
func parsePackage(lex *lexer) (string, error) {
	text := lex.text()
	if text != "package" {
		return "", fmt.Errorf("[BUG] not found package, text=%s", text)
	}
	lex.next()
	return lex.text(), nil
}

// "service var '{' serviceContent '}'
func parseService(lex *lexer) (*Service, error) {
	text := lex.text()
	if text != "service" {
		return nil, fmt.Errorf("[BUG] not found service, text=%s", text)
	}

	// メッセージ名を取得する {
	lex.next()
	name := lex.text()
	lex.next()
	// }

	// メッセージの中身を取得する {
	/// '{' を消費する {
	lex.next()
	/// }
	rpcs, err := parseServiceContent(lex)
	if err != nil {
		return nil, err
	}
	// }

	// '}' を消費する {
	lex.next()
	// }

	return &Service{
		Name: name,
		RPCs: rpcs,
	}, nil
}

// rpc
func parseServiceContent(lex *lexer) ([]*RPC, error) {
	var rpcs []*RPC
	for lex.text() != "}" {
		if lex.token != scanner.Comment {
			return nil, fmt.Errorf("not found comment, text=%s", lex.text())
		}
		comments := parseComments(lex)

		switch lex.text() {
		case "rpc":
			// rpc を消費する {
			lex.next()
			// }

			var rpc *RPC
			rpc, err := parseRPC(lex)
			if err != nil {
				return nil, err
			}
			rpc.Comments = append(rpc.Comments, comments...)
			rpcs = append(rpcs, rpc)
		default:
			return nil, fmt.Errorf("not found rpc, text=%s", lex.text())
		}
	}
	return rpcs, nil
}

// Name'('Argument')' 'returns' '('Return')' '{''}'
func parseRPC(lex *lexer) (*RPC, error) {
	rpc := &RPC{}

	for lex.text() != "}" {
		token := lex.text()
		if rpc.Name == "" {
			rpc.Name = token
			lex.next()
			continue
		}
		if rpc.Argument == nil {
			// '(' を消費する {
			lex.next()
			// }

			rpc.Argument = parseType(lex)

			// ')' を消費する {
			lex.next()
			// }
			continue
		}
		if rpc.Return == nil {
			// 'returns' を消費する {
			lex.next()
			// }
			// '(' を消費する {
			lex.next()
			// }

			rpc.Return = parseType(lex)
			lex.next()

			// ')' を消費する {
			lex.next()
			// }
			continue
		}
		// 消費する {
		lex.next()
		// }
	}

	// '}' を消費する {
	lex.next()
	// }

	return rpc, nil
}

// "message" var '{' messageContent '}'
func parseMessage(lex *lexer) (*Message, error) {
	text := lex.text()
	if text != "message" {
		return nil, fmt.Errorf("not found message, text=%s", text)
	}

	// メッセージ名を取得する {
	lex.next()
	name := lex.text()
	lex.next()
	// }

	// メッセージの中身を取得する {
	/// '{' を消費する {
	lex.next()
	/// }
	fields, nests, enums, oneofs, err := parseMessageContent(lex)
	if err != nil {
		return nil, err
	}
	// }

	// '}' を消費する {
	lex.next()
	// }

	return &Message{
		Name:   name,
		Fields: fields,
		Nests:  nests,
		Enums:  enums,
		Oneofs: oneofs,
	}, nil
}

// "message"
// "enum"
// field
func parseMessageContent(lex *lexer) (fields []*Field, messages []*Message, enums []*Enum, oneofs []*Oneof, err error) {
	for lex.text() != "}" {
		if lex.token != scanner.Comment {
			return nil, nil, nil, nil, fmt.Errorf("not found comment, text=%s", lex.text())
		}
		comments := parseComments(lex)

		switch lex.text() {
		case "message":
			message, parseErr := parseMessage(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			message.Comments = append(message.Comments, comments...)
			messages = append(messages, message)
		case "enum":
			enum, parseErr := parseEnum(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			enum.Comments = append(enum.Comments, comments...)
			enums = append(enums, enum)
		case "oneof":
			oneof, parseErr := parseOneof(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			oneof.Comments = append(oneof.Comments, comments...)
			oneofs = append(oneofs, oneof)
		default:
			field, parseErr := parseField(lex)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			field.Comments = append(field.Comments, comments...)
			fields = append(fields, field)
		}
	}

	return fields, messages, enums, oneofs, nil
}

// "enum" var '{' EnumContent '}'
func parseEnum(lex *lexer) (*Enum, error) {
	text := lex.text()
	if text != "enum" {
		return nil, fmt.Errorf("not found enum, text=%s", text)
	}

	// メッセージ名を取得する {
	lex.next()
	name := lex.text()
	lex.next()
	// }

	// メッセージの中身を取得する {
	/// '{' を消費する {
	lex.next()
	/// }
	fields, err := parseEnumContent(lex)
	if err != nil {
		return nil, err
	}
	// }

	// '}' を消費する {
	lex.next()
	// }

	return &Enum{
		Name:       name,
		EnumFields: fields,
	}, nil
}

// EnumField...}
func parseEnumContent(lex *lexer) ([]*EnumField, error) {
	var fields []*EnumField

	for lex.text() != "}" {
		field, err := parseEnumField(lex)
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}

	return fields, nil
}

// comment var '=' tag';'
func parseEnumField(lex *lexer) (*EnumField, error) {
	field := &EnumField{}

	// コメントを取得する {
	if lex.token != scanner.Comment {
		return nil, fmt.Errorf("not found comment, text=%s", lex.text())
	}
	field.Comments = parseComments(lex)
	// }

	field.Name = lex.text()

	// 残りを消費する {
	for lex.text() != ";" {
		lex.next()
	}
	lex.next()
	// }
	return field, nil
}

// "oneof" var '{' OneofContent '}'
func parseOneof(lex *lexer) (*Oneof, error) {
	text := lex.text()
	if text != "oneof" {
		return nil, fmt.Errorf("not found oneof, text=%s", text)
	}

	// 名前を取得する {
	lex.next()
	name := lex.text()
	lex.next()
	// }

	// 中身を取得する {
	/// '{' を消費する {
	lex.next()
	/// }
	fields, _, _, _, err := parseMessageContent(lex)
	if err != nil {
		return nil, err
	}
	// }

	// '}' を消費する {
	lex.next()
	// }

	return &Oneof{
		Name:   name,
		Fields: fields,
	}, nil
}

// type name = number validator';'
func parseField(lex *lexer) (*Field, error) {
	field := &Field{}

	for lex.text() != ";" {
		token := lex.text()
		if field.Type == nil {
			field.Type = parseType(lex)
			continue
		}
		if field.Name == "" {
			field.Name = token

			lex.next()
			continue
		}
		// 消費する {
		lex.next()
		// }
	}

	// ';' を消費する {
	lex.next()
	// }

	return field, nil
}

func parseType(lex *lexer) *Type {
	s := lex.text()
	lex.next()
	if s == "repeated" {
		t := parseType(lex)
		return &Type{
			Name:       t.Name,
			IsRepeated: true,
		}
	}
	for lex.text() == "." {
		s += lex.text()
		lex.next()
		s += lex.text()
		lex.next()
	}
	return &Type{
		Name:       s,
		IsRepeated: false,
	}
}

func parseComments(lex *lexer) []string {
	var s []string
	for lex.token == scanner.Comment {
		s = append(s, lex.text())
		lex.next()
	}
	return s
}
