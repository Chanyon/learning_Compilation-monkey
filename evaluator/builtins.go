package evaluator

import (
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltinByName("len"),
	"puts":  object.GetBuiltinByName("puts"),
	"first": object.GetBuiltinByName("first"),
	"last": object.GetBuiltinByName("last"),
	"rest": object.GetBuiltinByName("rest"),
	"push": object.GetBuiltinByName("push"),
	"shift": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got %d, want 2", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be an array, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.ELements)

			var newElements []object.Object
			newElements = append(newElements, args[1])
			if length == 0 {
				return &object.Array{ELements: newElements}
			} else {
				for i := 0; i < length; i++ {
					newElements = append(newElements, arr.ELements[i])
				}
			}
			return &object.Array{ELements: newElements}
		},
	},
	"remove": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got %d, want 1", len(args))
			}
			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `push` must be an array, got %s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.ELements)
			if length > 0 {
				newElements := make([]object.Object, length-1)
				copy(newElements, arr.ELements[0:length-1])
				return &object.Array{ELements: newElements}
			}
			return NULL
		},
	},
}
