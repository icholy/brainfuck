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

func formatIns(ins, params string, args ...interface{}) string {
	indent := strings.Repeat(" ", 8)
	params = fmt.Sprintf(params, args...)
	if params == "" {
		return indent + ins
	}
	ins = fmt.Sprintf("%-10s", ins)
	return indent + ins + params
}

func (c *Code) add(s string)                                { c.Instructions = append(c.Instructions, s) }
func (c *Code) addf(format string, args ...interface{})     { c.add(fmt.Sprintf(format, args...)) }
func (c *Code) Dir(format string, args ...interface{})      { c.addf(format, args...) }
func (c *Code) Label(name string)                           { c.addf("%s:", name) }
func (c *Code) Ins(ins, params string, args ...interface{}) { c.add(formatIns(ins, params, args...)) }

func CompileOp(op Op, code *Code) {
	switch op.Token {
	case GT:
		if op.Num == 1 {
			code.Ins("inc", "ebx")
		} else {
			code.Ins("add", "ebx, %d", op.Num)
		}
	case LT:
		if op.Num == 1 {
			code.Ins("dec", "ebx")
		} else {
			code.Ins("sub", "ebx, %d", op.Num)
		}
	case PLUS:
		if op.Num == 1 {
			code.Ins("inc", "byte [%s+ebx]", code.DataLabel())
		} else {
			code.Ins("add", "byte [%s+ebx], %d", code.DataLabel(), op.Num)
		}
	case SUB:
		if op.Num == 1 {
			code.Ins("dec", "byte [%s+ebx]", code.DataLabel())
		} else {
			code.Ins("sub", "byte [%s+ebx], %d", code.DataLabel(), op.Num)
		}
	case DOT:
		code.Ins("push", "dword [%s+ebx]", code.DataLabel())
		for i := 0; i < op.Num; i++ {
			code.Ins("call", "putchar")
		}
		code.Ins("pop", "ecx")
	case COMMA:
		for i := 0; i < op.Num; i++ {
			code.Ins("call", "getchar")
		}
		code.Ins("mov", "[%s+ebx], byte al", code.DataLabel())
	default:
		panic(fmt.Errorf("unsuported op: %s", op))
	}
}

func CompileLoop(loop Loop, code *Code) {
	start, end := code.LoopLabels()
	code.Label(start)
	code.Ins("mov", "al, [%s+ebx]", code.DataLabel())
	code.Ins("cmp", "al, 0")
	code.Ins("je", end)
	for _, n := range loop {
		CompileNode(n, code)
	}
	code.Ins("jmp", start)
	code.Label(end)
}

func CompileNode(node Node, code *Code) {
	switch node := node.(type) {
	case Op:
		CompileOp(node, code)
	case Loop:
		CompileLoop(node, code)
	default:
		panic(fmt.Errorf("unsuported node: %s", node))
	}
}

func Compile(p Program) []string {
	var code Code
	code.Dir("extern putchar, getchar")
	code.Dir("global main")
	code.Dir("segment .data")
	code.Dir("%s times %d db 0", code.DataLabel(), code.DataMax())
	code.Dir("segment .text")
	code.Label("main")
	code.Ins("enter", "0, 0")
	code.Ins("pusha", "")
	code.Ins("mov", "ebx, 0")
	for _, n := range p {
		CompileNode(n, &code)
	}
	code.Ins("popa", "")
	code.Ins("mov", "eax, 0")
	code.Ins("leave", "")
	code.Ins("ret", "")
	return code.Instructions
}
