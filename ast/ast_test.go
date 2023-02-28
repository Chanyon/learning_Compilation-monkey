package ast

import (
	"monkey/token"
	"testing"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}
	if program.String() != "let myVar = anotherVar;" {
		t.Errorf("program.String() wrong: got='%s'", program.String())
	}
}

func TestAssignStatement(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&ExpressionStatement{
				Token: token.Token{Type: token.IDENT, Literal: "foo"},
				Expression: &AssignExpression{
					Name: &Identifier{
						Token: token.Token{Type: token.IDENT, Literal: "foo"},
						Value: "foo",
					},
					Value: &IntegerLiteral{
						Token: token.Token{Type: token.INT, Literal: "6"},
						Value: 6,
					},
				},
			},
		},
	}
	if program.String() != "foo = 6;" {
		t.Errorf("program.String() wrong: got='%s'", program.String())
	}
}
