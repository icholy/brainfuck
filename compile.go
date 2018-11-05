package main

import (
	"fmt"
	"strings"
)

type Code struct {
	labels       int
	Instructions []string
}

func (c *Code) DataLabel() string { return "Data" }
func (c *Code) DataMax() int      { return 100000 }

func (c *Code) LoopLabels() (start, end string) {
	c.labels++
	return fmt.Sprintf("loop_%d", c.labels), fmt.Sprintf("end_loop_%d", c.labels)
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

func CompileOp(op Op, code *Code) error {
	switch op.Token {
	case GT:
		if op.Num == 1 {
			code.Op("inc ebx; %s", op)
		} else {
			code.Op("add ebx, %d; %s", op.Num, op)
		}
	case LT:
		if op.Num == 1 {
			code.Op("dec ebx; %s", op)
		} else {
			code.Op("sub ebx, %d; %s", op.Num, op)
		}
	case PLUS:
		if op.Num == 1 {
			code.Op("inc byte [%s+ebx]; %s", code.DataLabel(), op)
		} else {
			code.Op("add byte [%s+ebx], %d; %s", code.DataLabel(), op.Num, op)
		}
	case SUB:
		if op.Num == 1 {
			code.Op("dec byte [%s+ebx]; %s", code.DataLabel(), op)
		} else {
			code.Op("sub byte [%s+ebx], %d; %s", code.DataLabel(), op.Num, op)
		}
	case DOT:
		code.Op("push dword [%s+ebx]", code.DataLabel())
		for i := 0; i < op.Num; i++ {
			code.Op("call putchar")
		}
		code.Op("pop ecx")
	case COMMA:
		for i := 0; i < op.Num; i++ {
			code.Op("call getchar; ,")
		}
		code.Op("mov [%s+ebx], byte al", code.DataLabel())
	default:
		return fmt.Errorf("unsuported op: %s", op)
	}
	return nil
}

func CompileLoop(loop Loop, code *Code) error {
	start, end := code.LoopLabels()
	code.Ins("%s:", start)
	code.Op("mov al, [%s+ebx]", code.DataLabel())
	code.Op("cmp al, 0")
	code.Op("je %s", end)
	for _, n := range loop {
		if err := CompileNode(n, code); err != nil {
			return err
		}
	}
	code.Op("jmp %s; ]", start)
	code.Ins("%s:", end)
	return nil
}

func CompileSetup(code *Code) {
	code.Ins("extern putchar, getchar")
	code.Ins("global main")
	code.Ins("segment .data")
	code.Ins("%s times %d db 0", code.DataLabel(), code.DataMax())
	code.Ins("segment .text")
	code.Ins("main:")
	code.Op("enter 0, 0")
	code.Op("pusha")
	code.Op("mov ebx, 0")
}

func CompileCleanup(code *Code) {
	code.Op("popa")
	code.Op("mov eax, 0")
	code.Op("leave")
	code.Op("ret")
}

func CompileNode(node Node, code *Code) error {
	switch node := node.(type) {
	case Op:
		return CompileOp(node, code)
	case Loop:
		return CompileLoop(node, code)
	default:
		return fmt.Errorf("unsuported node: %s", node)
	}
	return nil
}

func Compile(p Program) ([]string, error) {
	var code Code
	CompileSetup(&code)
	for _, n := range p {
		if err := CompileNode(n, &code); err != nil {
			return nil, err
		}
	}
	CompileCleanup(&code)
	return code.Instructions, nil
}
