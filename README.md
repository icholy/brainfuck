# Brainfuck Compiler

> Compiles brainfuck programs to 32bit x86 nasm assembly

``` 

# compile
$ ./brainfuck -f examples/hanoi.bf -o hanoi.asm

# assemble
$ nasm -f elf hanoi.asm

# link
$ ld -m elf_i386 -o hanoi hanoi.o -lc
```
