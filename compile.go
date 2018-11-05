package main

import (
	"fmt"
	"strings"
)

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
