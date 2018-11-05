
mandelbrot: examples/mandelbrot.bf mandelbrot.asm mandelbrot.o
	./brainfuck -f examples/mandelbrot.bf -o mandelbrot.asm
	nasm -f elf32 mandelbrot.asm
	gcc -m32 -o mandelbrot mandelbrot.o
