package object

import (
	"fmt"
)

var Builtins = []struct {
	Name    string
	Builtin *Builtin
}{
	{
		"len",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got %d, want 1", len(args))
				}

				switch arg := args[0].(type) {
				case *String:
					return &Integer{Value: int64(len(arg.Value))}
				case *Array:
					return &Integer{Value: int64(len(arg.ELements))}
				default:
					return newError("argument to `len` not supported, got %s", args[0].Type())
				}
			},
		},
	},
	{
		"puts",
		&Builtin{
			Fn: func(args ...Object) Object {
				for _, arg := range args {
					fmt.Println(arg.Inspect())
				}
				return nil
			},
		},
	},
	{
		"first",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got %d, want 1", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `first` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				if len(arr.ELements) > 0 {
					return arr.ELements[0]
				}
				return nil
			},
		},
	},
	{
		"last",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got %d, want 1", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `last` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				length := len(arr.ELements) - 1
				if length >= 0 {
					return arr.ELements[length]
				}
				return nil
			},
		},
	},
	{
		"rest",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got %d, want 1", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `rest` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				length := len(arr.ELements)
				if length > 0 {
					newElements := make([]Object, length-1)
					copy(newElements, arr.ELements[1:length])
					return &Array{ELements: newElements}
				}
				return nil
			},
		},
	},
	{
		"push",
		&Builtin{
			Fn: func(args ...Object) Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got %d, want 2", len(args))
				}
				if args[0].Type() != ARRAY_OBJ {
					return newError("argument to `push` must be an array, got %s", args[0].Type())
				}

				arr := args[0].(*Array)
				length := len(arr.ELements)

				newElements := make([]Object, length+1)
				copy(newElements, arr.ELements)

				newElements[length] = args[1]

				return &Array{ELements: newElements}
			},
		},
	},
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}
