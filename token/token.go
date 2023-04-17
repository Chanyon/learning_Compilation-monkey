package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

// enum
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	// 标识符+字面量
	IDENT  = "IDENT"
	INT    = "INT"
	FLOAT  = "FLOAT"
	STRING = "STRING"

	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	SLASH    = "/"
	BANG     = "!"
	ASTERISK = "*"
	LT       = "<"
	LTQ      = "<="
	GT       = ">"
	GTQ      = ">="
	EQ       = "=="
	NOT_EQ   = "!="
	AND      = "&&"
	OR       = "||"
	// 分隔符
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"
	COLON     = ":"
	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	MACRO    = "MACRO"
	WHILE    = "WHILE"
	FOR      = "FOR"
	CLASS    = "CLASS"
	THIS     = "THIS"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
	"macro":  MACRO,
	"while":  WHILE,
	"for":    FOR,
	"class":  CLASS,
	"this":   THIS,
}

func NewToken(tokenType TokenType, ch byte) Token {
	return Token{Type: tokenType, Literal: string(ch)}
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
