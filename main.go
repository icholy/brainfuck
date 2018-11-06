package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	var infile, outfile string
	flag.StringVar(&infile, "f", "", "input file")
	flag.StringVar(&outfile, "o", "out.asm", "output file")
	flag.Parse()
	src, err := ioutil.ReadFile(infile)
	if err != nil {
		log.Fatal(err)
	}
	tokens := Tokenize(string(src))
	program, err := Parse(tokens)
	if err != nil {
		log.Fatal(err)
	}
	instructions := Compile(program)
	assembly := strings.Join(instructions, "\n")
	if err := ioutil.WriteFile(outfile, []byte(assembly), os.ModePerm); err != nil {
		log.Fatal(err)
	}
}
