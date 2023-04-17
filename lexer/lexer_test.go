package lexer

import (
	"monkey/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
	let ten = 10; 
	let add = fn(x, y) { 
			x + y; 
	}; 
	let result = add(five, ten); 
	!-/*+;
	5<10>5;
	if(5<10){
		return true;
	}else {
		return false;
	}
	10==10;
	10!=9;
	"foo";
	"bar";
	"foo bar";
	"test"e,";
	"test2\'\'";
	"te\n\rst";
	"test3";q;";
	"test"";
	"test";
	"test";,";
	[1,2];
	{"foo":"bar"};
	macro(x,y){x+y;};
	while;
	foo = 1;
	for;
	<=;
	>=;
	&&;
	||;
	class
	this
	`
	// "1.123";
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.PLUS, "+"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foo"},
		{token.SEMICOLON, ";"},
		{token.STRING, "bar"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foo bar"},
		{token.SEMICOLON, ";"},
		{token.STRING, "test\"e,"},
		{token.SEMICOLON, ";"},
		{token.STRING, "test2\\'\\'"},
		{token.SEMICOLON, ";"},
		{token.STRING, "te\\n\\rst"},
		{token.SEMICOLON, ";"},
		{token.STRING, "test3\";q;"},
		{token.SEMICOLON, ";"},
		{token.STRING, "test\""},
		{token.SEMICOLON, ";"},
		{token.STRING, "test"},
		{token.SEMICOLON, ";"},
		// {token.STRING, "1.123"},
		// {token.SEMICOLON, ";"},
		{token.STRING, "test\";,"},
		{token.SEMICOLON, ";"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		// macro(x,y){x+y;};
		{token.MACRO, "macro"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.WHILE, "while"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "foo"},
		{token.ASSIGN, "="},
		{token.INT, "1"},
		{token.SEMICOLON, ";"},
		{token.FOR, "for"},
		{token.SEMICOLON, ";"},
		{token.LTQ, "<="},
		{token.SEMICOLON, ";"},
		{token.GTQ, ">="},
		{token.SEMICOLON, ";"},
		{token.AND, "&&"},
		{token.SEMICOLON, ";"},
		{token.OR, "||"},
		{token.SEMICOLON, ";"},
		{token.CLASS, "class"},
		{token.THIS, "this"},
		{token.EOF, ""},
	}
	l := New(input)
	for i, tt := range tests {
		token := l.NextToken()
		if token.Type != tt.expectedType {
			t.Fatalf("tests[%d] - token type wrong. expected=%q, got=(%q)", i, tt.expectedType, token.Type)
		}
		if token.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=(%q)", i, tt.expectedLiteral, token.Literal)
		}
	}
}
