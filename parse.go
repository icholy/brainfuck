package main

import (
	"fmt"
	"strings"
)

type Node interface{ fmt.Stringer }

type Op struct {
	Token Token
	Num   int
}

func (o Op) String() string {
	if o.Num == 0 {
		return string(o.Token)
	}
	return fmt.Sprintf("(%s:%d)", o.Token, o.Num)
}

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

func ParseOp(pos int, tokens []Token) (Op, int, error) {
	op := Op{Token: tokens[pos]}
	for pos = pos; pos < len(tokens); pos++ {
		if tokens[pos] == op.Token {
			op.Num++
		} else {
			pos--
			break
		}
	}
	return op, pos, nil
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
			onode, opos, err := ParseOp(pos, tokens)
			if err != nil {
				return nil, opos, err
			}
			pos = opos
			nodes = append(nodes, onode)
		}
	}
	return nil, 0, fmt.Errorf("missing %s", EOF)
}

func Parse(tokens []Token) (Program, error) {
	nodes, _, err := ParseNodes(0, tokens)
	return Program(nodes), err
}
