package ast

import (
	"testing"

	"gocompiler/token"
)

func TestString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.Let, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.Identifier, Literal: "kek"},
					Value: "kek",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.Identifier, Literal: "lol"},
					Value: "lol",
				},
			},
		},
	}

	if program.String() != "let kek = lol;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
