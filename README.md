# Brainfuck Compiler

> Compiles brainfuck programs to 32bit x86 [nasm](https://www.nasm.us/) assembly

``` sh
# compile
$ ./brainfuck -f examples/hello.bf -o hello.asm

# assemble
$ nasm -f elf32 hello.asm

# link
$ gcc -m32 -o hello hello.o
```
