package main

type Token string

func (t Token) String() string { return string(t) }

const (
	GT      = Token("GT")
	LT      = Token("LT")
	PLUS    = Token("PLUS")
	SUB     = Token("SUB")
	DOT     = Token("DOT")
	COMMA   = Token("COMMA")
	OPEN    = Token("OPEN")
	CLOSE   = Token("CLOSE")
	EOF     = Token("EOF")
	INVALID = Token("INVALID")
)

var mapping = map[rune]Token{
	'>': GT,
	'<': LT,
	'+': PLUS,
	'-': SUB,
	'.': DOT,
	'[': OPEN,
	']': CLOSE,
	',': COMMA,
}

func Tokenize(s string) []Token {
	var tokens []Token
	for _, r := range s {
		switch r {
		case ' ', '\n', '\r', '\t':
		default:
			tok, ok := mapping[r]
			if !ok {
				tok = INVALID
			}
			tokens = append(tokens, tok)
		}
	}
	return append(tokens, EOF)
}
