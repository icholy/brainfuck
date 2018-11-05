
all: mandelbrot life hanoi fib hello

mandelbrot:
	./brainfuck -f examples/mandelbrot.bf -o mandelbrot.asm
	nasm -f elf32 mandelbrot.asm
	gcc -m32 -o mandelbrot mandelbrot.o

life:
	./brainfuck -f examples/life.bf -o life.asm
	nasm -f elf32 life.asm
	gcc -m32 -o life life.o

hanoi:
	./brainfuck -f examples/hanoi.bf -o hanoi.asm
	nasm -f elf32 hanoi.asm
	gcc -m32 -o hanoi hanoi.o

fib:
	./brainfuck -f examples/fib.bf -o fib.asm
	nasm -f elf32 fib.asm
	gcc -m32 -o fib fib.o

hello:
	./brainfuck -f examples/hello.bf -o hello.asm
	nasm -f elf32 hello.asm
	gcc -m32 -o hello hello.o

clean:
	rm -f mandelbrot* life* hanoi* fib* hello*
