package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	Illegal = "Illegal"
	EOF     = "EOF"

	// Identifiers + Literals
	Identifier = "Identifier"
	Int        = "Int"
	String     = "String"

	// Operators
	Assign   = "="
	Plus     = "+"
	Minus    = "-"
	Bang     = "!"
	Asterisk = "*"
	Slash    = "/"
	Equal    = "=="
	NotEqual = "!="

	GreaterThan = ">"
	LessThan    = "<"

	// Delimiters
	Comma     = ","
	Semicolon = ";"
	Colon     = ":"

	LeftParen    = "("
	RightParen   = ")"
	LeftBrace    = "{"
	RightBrace   = "}"
	LeftBracket  = "["
	RightBracket = "]"

	// Keywords
	Function = "Function"
	Let      = "Let"
	True     = "True"
	False    = "False"
	If       = "If"
	Else     = "Else"
	Return   = "Return"
)

var keywords = map[string]TokenType{
	"function": Function,
	"let":      Let,
	"return":   Return,
	"true":     True,
	"false":    False,
	"if":       If,
	"else":     Else,
}

func LookupIdentifierType(identifier string) TokenType {
	tok, ok := keywords[identifier]
	if !ok {
		return Identifier
	}

	return tok
}
