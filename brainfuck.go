package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

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

type Node interface{ fmt.Stringer }

type Op Token

func (o Op) String() string { return string(o) }

type Loop []Node

func (l Loop) String() string {
	nodes := make([]string, len(l))
	for i, n := range l {
		nodes[i] = n.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(nodes, ", "))
}

type Program []Node

func (p Program) String() string {
	nodes := make([]string, len(p))
	for i, n := range p {
		nodes[i] = n.String()
	}
	return strings.Join(nodes, ", ")
}

func ParseNodes(pos int, tokens []Token) ([]Node, int, error) {
	var nodes []Node
	for pos = pos; pos < len(tokens); pos++ {
		tok := tokens[pos]
		switch tok {
		case INVALID:
			// ignore invalid tokens
		case EOF:
			return nodes, pos, nil
		case CLOSE:
			return nodes, pos, nil
		case OPEN:
			lnodes, lpos, err := ParseNodes(pos+1, tokens)
			if err != nil {
				return nil, lpos, err
			}
			if tok := tokens[lpos]; tok != CLOSE {
				return nil, pos, fmt.Errorf("expected %s, got %s", CLOSE, tok)
			}
			pos = lpos
			nodes = append(nodes, Loop(lnodes))
		default:
			nodes = append(nodes, Op(tok))
		}
	}
	return nil, 0, fmt.Errorf("missing %s", EOF)
}

func Parse(tokens []Token) (Program, error) {
	nodes, _, err := ParseNodes(0, tokens)
	return Program(nodes), err
}

type Labels struct {
	count int
}

func (l *Labels) Data() string  { return "Data" }
func (l *Labels) Index() string { return "Index" }

func (l *Labels) Loop() (start, end string) {
	l.count++
	return fmt.Sprintf("loop_%d", l.count), fmt.Sprintf("end_loop_%d", l.count)
}

func CompileOp(op Op, labels *Labels) ([]string, error) {
	switch Token(op) {
	case GT:
		return []string{
			fmtOp("inc dword [%s]; >", labels.Index()),
		}, nil
	case LT:
		return []string{
			fmtOp("dec dword [%s]; <", labels.Index()),
		}, nil
	case PLUS:
		return []string{
			fmtOp("mov eax, [%s]; +", labels.Index()),
			fmtOp("inc byte [%s+eax]", labels.Data()),
		}, nil
	case SUB:
		return []string{
			fmtOp("mov eax, [%s]; -", labels.Index()),
			fmtOp("dec byte [%s+eax]", labels.Data()),
		}, nil
	case DOT:
		return []string{
			fmtOp("xor eax, eax; ."),
			fmtOp("mov al, [%s]", labels.Index()),
			fmtOp("push dword [%s+eax]", labels.Data()),
			fmtOp("call putchar"),
			fmtOp("pop ecx"),
		}, nil
	case COMMA:
		return []string{
			fmtOp("call getch; ,"),
			fmtOp("mov ebx, [%s]", labels.Index()),
			fmtOp("mov [%s+ebx], byte al", labels.Data()),
		}, nil
	default:
		return nil, fmt.Errorf("unsuported op: %s", op)
	}
}

func CompileLoop(loop Loop, labels *Labels) ([]string, error) {
	start, end := labels.Loop()
	instructions := []string{
		fmtIns("%s:", start),
		fmtOp("mov eax, [%s]; [", labels.Index()),
		fmtOp("mov al, [%s+eax]", labels.Data()),
		fmtOp("cmp al, 0"),
		fmtOp("je %s", end),
	}
	for _, n := range loop {
		ins, err := CompileNode(n, labels)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, ins...)
	}
	return append(instructions,
		fmtOp("jmp %s; ]", start),
		fmtIns("%s:", end),
	), nil
}

func CompileSetup(labels *Labels) []string {
	return []string{
		fmtIns("extern putchar, getch"),
		fmtIns("global main"),
		fmtIns("segment .data"),
		fmtIns("%s times 100000 db 0", labels.Data()),
		fmtIns("%s dd 0", labels.Index()),
		fmtIns("segment .text"),
		fmtIns("main:"),
		fmtOp("enter 0, 0"),
		fmtOp("pusha"),
	}
}

func CompileCleanup(labels *Labels) []string {
	return []string{
		fmtOp("popa"),
		fmtOp("mov eax, 0"),
		fmtOp("leave"),
		fmtOp("ret"),
	}
}


func fmtIns(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func fmtOp(format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args...)
	parts := strings.SplitN(s, " ", 2)
	s = fmt.Sprintf("\t%s \t\t%s", parts[0], strings.Join(parts[1:], " "))
	parts = strings.SplitN(s, ";", 2)
	if len(parts) == 1 {
		return parts[0]
	}
	return fmt.Sprintf("%s\t\t\t;%s", parts[0], strings.Join(parts[1:], ";"))
}

func CompileNode(node Node, labels *Labels) ([]string, error) {
	var instructions []string
	switch node := node.(type) {
	case Op:
		ins, err := CompileOp(node, labels)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, ins...)
	case Loop:
		ins, err := CompileLoop(node, labels)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, ins...)
	case Program:
		instructions = append(instructions, CompileSetup(labels)...)
		for _, n := range node {
			ins, err := CompileNode(n, labels)
			if err != nil {
				return nil, err
			}
			instructions = append(instructions, ins...)
		}
		instructions = append(instructions, CompileCleanup(labels)...)
	default:
		return nil, fmt.Errorf("unsuported node: %s", node)
	}
	return instructions, nil
}

func Compile(p Program) ([]string, error) {
	var labels Labels
	return CompileNode(p, &labels)
}

func main() {
	var infile, outfile string
	flag.StringVar(&infile, "f", "", "input file")
	flag.StringVar(&outfile, "o", "out.asm", "output file")
	flag.Parse()
	src, err := ioutil.ReadFile(infile)
	if err != nil {
		log.Fatal(err)
	}
	tokens := Tokenize(string(src))
	program, err := Parse(tokens)
	if err != nil {
		log.Fatal(err)
	}
	instructions, err := Compile(program)
	if err != nil {
		log.Fatal(err)
	}
	assembly := strings.Join(instructions, "\n")
	if err := ioutil.WriteFile(outfile, []byte(assembly), os.ModePerm); err != nil {
		log.Fatal(err)
	}
}
