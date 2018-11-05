
mandelbrot:
	./brainfuck -f examples/mandelbrot.bf -o mandelbrot.asm
	nasm -f elf32 mandelbrot.asm
	gcc -m32 -o mandelbrot mandelbrot.o

life:
	./brainfuck -f examples/life.bf -o life.asm
	nasm -f elf32 life.asm
	gcc -m32 -o life life.o


clean:
	rm -f mandelbrot* life*
