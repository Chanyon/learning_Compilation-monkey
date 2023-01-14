package repl

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"monkey/ast"
	"monkey/compiler"
	"monkey/evaluator"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
)

const PROMPT = ">>"
const MONKEY_FACE = ` __,__
    .--. .-" "-. .--.
  / .. \/ .-. .-. \/ .. \
    | | '| / Y \ |' | |
    | \ \ \ 0 | 0 / / / |
  \ '- ,\.-"""""""-./, -' /
  ''-' /_ ^ ^ _\ '-''
        | \._ _./ |
        \ \ '~' / /
        '._'-=-' _'
          '-----'
`

func StartVM(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)

		p := parser.New(l)
		program := p.ParserProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)

		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n%s\n", err)
			continue
		}

		code := comp.ByteCode()
		constants = code.Constants
		machine := vm.NewWithGlobalStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n%s\n", err)
			continue
		}
		stackTop := machine.LastPoppedStackElem()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")

	}
}

func StartEval(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment() // macro environment

	for {
		fmt.Fprint(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		l := lexer.New(line)

		p := parser.New(l)
		program := p.ParserProgram()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluator.DefineMacro(program, macroEnv)
		expanded := evaluator.ExpandMacros(program, macroEnv)

		evaluated := evaluator.Eval(expanded, env)

		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func StartFile(filePath string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("reading file error: %s", err)
	}

	program := parse(string(data))
	compiler := compiler.New()

	err = compiler.Compile(program)

	if err != nil {
		fmt.Printf("compiler error: %s", err)
	}

	vm := vm.New(compiler.ByteCode())
	err = vm.Run()

	if err != nil {
		fmt.Printf("vm run failed: %s", err)
	}

		stackElem := vm.LastPoppedStackElem()
		fmt.Println(stackElem.Inspect())
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParserProgram()
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
