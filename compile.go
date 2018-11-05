package main

import (
	"fmt"
	"strings"
)

type Code struct {
	Instructions []string
}

func (c *Code) Ins(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	c.Instructions = append(c.Instructions, s)
}

func (c *Code) Op(format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	parts := strings.SplitN(s, " ", 2)
	s = fmt.Sprintf("\t%s \t\t%s", parts[0], strings.Join(parts[1:], " "))
	parts = strings.SplitN(s, ";", 2)
	if len(parts) == 1 {
		s = parts[0]
	} else {
		s = fmt.Sprintf("%s\t\t\t;%s", parts[0], strings.Join(parts[1:], ";"))
	}
	c.Instructions = append(c.Instructions, s)
}

type Labels struct {
	count int
}

func (l *Labels) Data() string { return "Data" }

func (l *Labels) Loop() (start, end string) {
	l.count++
	return fmt.Sprintf("loop_%d", l.count), fmt.Sprintf("end_loop_%d", l.count)
}

func CompileOp(op Op, code *Code, labels *Labels) error {
	switch op.Token {
	case GT:
		code.Op("add ebx, %d; %s", op.Num, op)
	case LT:
		code.Op("sub ebx, %d; %s", op.Num, op)
	case PLUS:
		code.Op("add byte [%s+ebx], %d", labels.Data(), op.Num)
	case SUB:
		code.Op("sub byte [%s+ebx], %d", labels.Data(), op.Num)
	case DOT:
		code.Op("push dword [%s+ebx]", labels.Data())
		for i := 0; i < op.Num; i++ {
			code.Op("call putchar")
		}
		code.Op("pop ecx")
	case COMMA:
		for i := 0; i < op.Num; i++ {
			code.Op("call getch; ,")
		}
		code.Op("mov [%s+ebx], byte al", labels.Data())
	default:
		return fmt.Errorf("unsuported op: %s", op)
	}
	return nil
}

func CompileLoop(loop Loop, code *Code, labels *Labels) error {
	start, end := labels.Loop()
	code.Ins("%s:", start)
	code.Op("mov al, [%s+ebx]", labels.Data())
	code.Op("cmp al, 0")
	code.Op("je %s", end)
	for _, n := range loop {
		if err := CompileNode(n, code, labels); err != nil {
			return err
		}
	}
	code.Op("jmp %s; ]", start)
	code.Ins("%s:", end)
	return nil
}

func CompileSetup(code *Code, labels *Labels) {
	code.Ins("extern putchar, getch")
	code.Ins("global main")
	code.Ins("segment .data")
	code.Ins("%s times 100000 db 0", labels.Data())
	code.Ins("segment .text")
	code.Ins("main:")
	code.Op("enter 0, 0")
	code.Op("pusha")
	code.Op("mov ebx, 0")
}

func CompileCleanup(code *Code, labels *Labels) {
	code.Op("popa")
	code.Op("mov eax, 0")
	code.Op("leave")
	code.Op("ret")
}

func CompileNode(node Node, code *Code, labels *Labels) error {
	switch node := node.(type) {
	case Op:
		return CompileOp(node, code, labels)
	case Loop:
		return CompileLoop(node, code, labels)
	case Program:
		CompileSetup(code, labels)
		for _, n := range node {
			if err := CompileNode(n, code, labels); err != nil {
				return err
			}
		}
		CompileCleanup(code, labels)
	default:
		return fmt.Errorf("unsuported node: %s", node)
	}
	return nil
}

func Compile(p Program) ([]string, error) {
	var (
		labels Labels
		code   Code
	)
	if err := CompileNode(p, &code, &labels); err != nil {
		return nil, err
	}
	return code.Instructions, nil
}
