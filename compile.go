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
	switch op.Token {
	case GT:
		return []string{
			fmtOp("add dword [%s], %d; %s", labels.Index(), op.Num, op),
		}, nil
	case LT:
		return []string{
			fmtOp("sub dword [%s], %d; %s", labels.Index(), op.Num, op),
		}, nil
	case PLUS:
		return []string{
			fmtOp("mov eax, [%s]; %s", labels.Index(), op),
			fmtOp("add byte [%s+eax], %d", labels.Data(), op.Num),
		}, nil
	case SUB:
		return []string{
			fmtOp("mov eax, [%s]; %s", labels.Index(), op),
			fmtOp("sub byte [%s+eax], %d", labels.Data(), op.Num),
		}, nil
	case DOT:
		instructions := []string{
			fmtOp("xor eax, eax; ."),
			fmtOp("mov al, [%s]", labels.Index()),
			fmtOp("push dword [%s+eax]", labels.Data()),
		}
		for i := 0; i < op.Num; i++ {
			instructions = append(instructions,
				fmtOp("call putchar"),
			)
		}
		instructions = append(instructions,
			fmtOp("pop ecx"),
		)
		return instructions, nil
	case COMMA:
		var instructions []string
		for i := 0; i < op.Num; i++ {
			instructions = append(instructions,
				fmtOp("call getch; ,"),
			)
		}
		instructions = append(instructions,
			fmtOp("mov ebx, [%s]", labels.Index()),
			fmtOp("mov [%s+ebx], byte al", labels.Data()),
		)
		return instructions, nil
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
