package vm

import (
	"fmt"
	"monkey/code"
	"monkey/compiler"
	"monkey/object"
)

// 虚拟机需要： 指令集合、常量池、栈 ...

const StackSize = 1 << 11
const GlobalSize = 1 << 16
const MaxFrames = 1024

// 全局的bool值
var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}

var Null = &object.Null{}

type VM struct {
	constants []object.Object
	// instructions code.Instruction
	stack      []object.Object
	sp         uint //指向栈中下一个空闲槽
	globals    []object.Object
	frames     []*Frame
	frameIndex int
}

// stack frame 函数调用栈
type Frame struct {
	closureFn   *object.Closure
	ip          int
	basePointer uint
}

func NewFrame(fn *object.Closure, basePointer uint) *Frame {
	return &Frame{closureFn: fn, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instruction {
	return f.closureFn.Fn.Instructions
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.frameIndex-1]
}
func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.frameIndex] = f
	vm.frameIndex += 1
}
func (vm *VM) popFrame() *Frame {
	vm.frameIndex -= 1
	return vm.frames[vm.frameIndex]
}

// create VM
func New(byteCode *compiler.ByteCode) *VM {
	mainFn := &object.CompiledFunction{Instructions: byteCode.Instruction}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		// instructions: byteCode.Instruction,
		constants:  byteCode.Constants,
		stack:      make([]object.Object, StackSize),
		sp:         0,
		globals:    make([]object.Object, GlobalSize),
		frames:     frames,
		frameIndex: 1,
	}
}

func NewWithGlobalStore(byteCode *compiler.ByteCode, globals []object.Object) *VM {
	vm := New(byteCode)
	vm.globals = globals
	return vm
}

// return stack top element
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// 取指->解码->循环执行
func (vm *VM) Run() error {
	var ip int
	var ins code.Instruction
	var op code.Opcode
	// length := len(ins)

	// for ip := 0; ip < length; ip++ {
	currentFrameLen := len(vm.currentFrame().Instructions()) - 1
	for vm.currentFrame().ip < currentFrameLen {
		vm.currentFrame().ip += 1

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		// op := code.Opcode(ins[ip])
		op = code.Opcode(ins[ip])

		// 解码
		switch op {
		case code.OpConstant:
			// constIndex := code.ReadUnit16(c.instruction[ip+1:])
			constIndex := code.ReadUnit16(ins[ip+1:])
			// ip += 2
			vm.currentFrame().ip += 2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpPop:
			_ = vm.pop()

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpLessThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}
		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpAnd:
			pos := int(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
				vm.push(condition)
			}
		case code.OpOr:
			pos := int(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				// skip Jump指令
				vm.currentFrame().ip = pos - 1
			} else {
				// 如果值为真， 重新入栈
				vm.push(condition)
			}
		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			globalIndex := code.ReadUnit16(ins[ip+1:])
			vm.currentFrame().ip += 2

			vm.globals[globalIndex] = vm.pop()
		case code.OpGetGlobal:
			globalIndex := code.ReadUnit16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.globals[globalIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numElements := uint(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numElements := uint(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}

			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}
			// fn, ok := vm.stack[vm.sp-1].(*object.CompiledFunction)

		case code.OpReturnValue:
			returnValue := vm.pop()
			frame := vm.popFrame() //* 回到mainFn
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1
			// _ = vm.pop() //pop object.CompiledFn

			err := vm.push(Null)
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+uint(localIndex)] = vm.pop()
		case code.OpGetLocal:
			localIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+uint(localIndex)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constantIndex := code.ReadUnit16(ins[ip+1:])
			numFreeVar := code.ReadUnit8(ins[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(constantIndex, uint16(numFreeVar))
			if err != nil {
				return err
			}
		case code.OpGetFreeVar:
			freeIndex := code.ReadUnit8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().closureFn
			err := vm.push(currentClosure.FreeVar[freeIndex])
			if err != nil {
				return err
			}
		case code.OpCurrnetClosure:
			currentClosure := vm.currentFrame().closureFn
			err := vm.push(currentClosure)

			if err != nil {
				return err
			}
		case code.OpLoop:
			pos := int(code.ReadUnit16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1
		} //switch end
		currentFrameLen = len(vm.currentFrame().Instructions()) - 1
	}
	return nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}
	vm.stack[vm.sp] = obj
	vm.sp += 1
	return nil
}

func (vm *VM) pop() object.Object {
	// fmt.Println("vm.pop() -> vm.sp: ", vm.sp)
	obj := vm.stack[vm.sp-1]
	vm.sp -= 1
	return obj
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()
	leftType := left.Type()
	rightType := right.Type()

	if leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}
	if leftType == object.STRING && rightType == object.STRING {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %q %q", leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operation: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value
	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBoolObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBoolObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value
	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBoolObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBoolObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBoolObject(leftValue > rightValue))
	case code.OpLessThan:
		return vm.push(nativeBoolToBoolObject(leftValue < rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBoolObject(input bool) *object.Boolean {
	if input {
		return True
	} else {
		return False
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != object.INTEGER_OBJ {
		return fmt.Errorf("unsupported type of negation: %s", operand.Type())
	}
	value := operand.(*object.Integer).Value
	return vm.push(&object.Integer{Value: -value})
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	default:
		return true
	}
}

// * 虚拟机构建数组字面量
func (vm *VM) buildArray(startIndex, endIndex uint) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{ELements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex uint) (object.Object, error) {
	hashElements := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		val := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: val}

		hashKey, ok := key.(object.HashAble)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashElements[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashElements}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	if left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ {
		return vm.executeArrayIndex(left, index)
	} else if left.Type() == object.HASH_OBJ {
		return vm.executeHashIndex(left, index)
	} else {
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(left, index object.Object) error {
	array := left.(*object.Array)
	idx := index.(*object.Integer).Value

	maxLen := int64(len(array.ELements) - 1)
	if idx < 0 || idx > maxLen {
		return vm.push(Null)
	}

	return vm.push(array.ELements[idx])
}

func (vm *VM) executeHashIndex(left, index object.Object) error {
	hash := left.(*object.Hash)
	key, ok := index.(object.HashAble)

	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hash.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeCall(numArgs uint8) error {
	callFn := vm.stack[vm.sp-uint(numArgs)-1]
	switch callType := callFn.(type) {
	case *object.Closure:
		return vm.callFunction(callType, numArgs)
	case *object.Builtin:
		return vm.Builtin(callType, numArgs)
	default:
		return fmt.Errorf("calling non-closure-function and non-built-in")
	}
}

func (vm *VM) callFunction(clFn *object.Closure, numArgs uint8) error {
	//!
	if clFn.Fn.NumParameters != int(numArgs) {
		return fmt.Errorf("wrong number of arguments.want=%d, got=%d",
			clFn.Fn.NumParameters, numArgs)
	}
	frame := NewFrame(clFn, vm.sp-uint(numArgs))
	vm.pushFrame(frame)

	vm.sp = frame.basePointer + uint(clFn.Fn.NumLocals)

	return nil
}

func (vm *VM) Builtin(builtin *object.Builtin, numArgs uint8) error {
	arguments := vm.stack[vm.sp-uint(numArgs) : vm.sp]
	result := builtin.Fn(arguments...)

	vm.sp = vm.sp - uint(numArgs)

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) pushClosure(constantIdx uint16, numFree uint16) error {
	constant := vm.constants[constantIdx]
	fn, ok := constant.(*object.CompiledFunction)

	if !ok {
		return fmt.Errorf("not a function: %v", constant)
	}

	free := make([]object.Object, numFree)
	var i uint
	for i = 0; i < uint(numFree); i++ {
		free[i] = vm.stack[vm.sp-uint(numFree)+i]
	}

	closure := &object.Closure{Fn: fn, FreeVar: free}
	return vm.push(closure)
}
