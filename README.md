# Brainfuck Compiler

> Compiles brainfuck programs to 32bit x86 nasm assembly

``` 

# compile
$ ./brainfuck -f examples/hanoi.bf -o hanoi.asm

# assemble
$ nasm -f elf hanoi.asm

# link
$ gcc -m32 -o hanoi hanoi.o
```
