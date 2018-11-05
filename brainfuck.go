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
			fmt.Sprintf("inc dword [%s]", labels.Index()),
		}, nil
	case LT:
		return []string{
			fmt.Sprintf("dec dword [%s]", labels.Index()),
		}, nil
	case PLUS:
		return []string{
			fmt.Sprintf("mov eax, [%s]", labels.Index()),
			fmt.Sprintf("inc byte [%s+eax]", labels.Data()),
		}, nil
	case SUB:
		return []string{
			fmt.Sprintf("mov eax, [%s]", labels.Index()),
			fmt.Sprintf("dec byte [%s+eax]", labels.Data()),
		}, nil
	case DOT:
		return []string{
			fmt.Sprintf("xor eax, eax"),
			fmt.Sprintf("mov al, [%s]", labels.Index()),
			fmt.Sprintf("push dword [%s+eax]", labels.Data()),
			fmt.Sprintf("call _putchar"),
			fmt.Sprintf("pop ecx"),
		}, nil
	case COMMA:
		return []string{
			fmt.Sprintf("call _getch"),
			fmt.Sprintf("mov ebx, [%s]", labels.Index()),
			fmt.Sprintf("mov [%s+ebx], byte al", labels.Data()),
		}, nil
	default:
		return nil, fmt.Errorf("unsuported op: %s", op)
	}
}

func CompileLoop(loop Loop, labels *Labels) ([]string, error) {
	start, end := labels.Loop()
	instructions := []string{
		fmt.Sprintf("%s:", start),
		fmt.Sprintf("mov eax, [%s]", labels.Index()),
		fmt.Sprintf("mov al, [%s+eax]", labels.Data()),
		fmt.Sprintf("cmp al, 0"),
		fmt.Sprintf("je %s", end),
	}
	for _, n := range loop {
		ins, err := CompileNode(n, labels)
		if err != nil {
			return nil, err
		}
		instructions = append(instructions, ins...)
	}
	return append(instructions,
		fmt.Sprintf("jmp %s", start),
		fmt.Sprintf("%s:", end),
	), nil
}

func CompileSetup(labels *Labels) []string {
	return []string{
		fmt.Sprintf("segment .data"),
		fmt.Sprintf("%s times 100000 db 0", labels.Data()),
		fmt.Sprintf("%s dd 0", labels.Index()),
		fmt.Sprintf("segment .text"),
		fmt.Sprintf("extern _putchar, _getch"),
		fmt.Sprintf("global _asm_main"),
		fmt.Sprintf("_asm_main:"),
		fmt.Sprintf("enter 0, 0"),
		fmt.Sprintf("pusha"),
	}
}

func CompileTeardown(labels *Labels) []string {
	return []string{
		fmt.Sprintf("popa"),
		fmt.Sprintf("mov eax, 0"),
		fmt.Sprintf("leave"),
		fmt.Sprintf("ret"),
	}
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
		instructions = append(instructions, CompileTeardown(labels)...)
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
