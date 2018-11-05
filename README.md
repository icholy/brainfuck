# Brainfuck Compiler

> Compiles brainfuck programs to 32bit x86 [nasm](https://www.nasm.us/) assembly

``` sh
# compile
$ ./brainfuck -f examples/mandelbrot.bf -o mandelbrot.asm

# assemble
$ nasm -f elf32 mandelbrot.asm

# link
$ gcc -m32 -o mandelbrot mandelbrot.o
```
