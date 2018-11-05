package main

import (
	"fmt"
	"strings"
)

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
