package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instruction []byte //字节码集合

func (ins Instruction) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}

		operands, readByteLength := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n\t", i, ins.fmtInstruction(def, operands))

		i += readByteLength + 1
	}

	return out.String()
}

func (ins Instruction) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("Error: operand len %d does not match defined %d\n", len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("Error: unhandled operandCount for  %s\n", def.Name)
}

type Opcode byte

// ! 操作码
const (
	OpConstant Opcode = iota
	OpAdd
	OpPop
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual       // ==
	OpNotEqual    // !=
	OpGreaterThan // > <
	// OpLessThan <
	OpMinus // -
	OpBang  // !
	OpJump  // jump
	OpJumpNotTruthy
	OpNull
	OpSetGlobal
	OpGetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturnValue
	OpReturn
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpClosure
	OpGetFreeVar
	OpCurrnetClosure
	OpLoop
)

type Definition struct {
	Name          string //操作码名称
	OperandWidths []int  //每个操作数占用的字节数
}

// ! 不要漏OpCode
var definitions = map[Opcode]*Definition{
	OpConstant:       {"OpConstant", []int{2}},
	OpAdd:            {"OpAdd", []int{}},
	OpPop:            {"OpPop", []int{}},
	OpSub:            {"OpSub", []int{}},
	OpMul:            {"OpMul", []int{}},
	OpDiv:            {"OpDiv", []int{}},
	OpTrue:           {"OpTrue", []int{}},
	OpFalse:          {"OpFalse", []int{}},
	OpEqual:          {"OpEqual", []int{}},
	OpNotEqual:       {"OpNotEqual", []int{}},
	OpGreaterThan:    {"OpGreaterThan", []int{}},
	OpMinus:          {"OpMinus", []int{}},
	OpBang:           {"OpBang", []int{}},
	OpJumpNotTruthy:  {"OpJumpNotTruthy", []int{2}},
	OpJump:           {"OpJump", []int{2}},
	OpNull:           {"OpNull", []int{}},
	OpSetGlobal:      {"OpSetGlobal", []int{2}},
	OpGetGlobal:      {"OpGetGlobal", []int{2}},
	OpArray:          {"OpArray", []int{2}}, //! 65535个数组元素 u16
	OpHash:           {"OpHash", []int{2}},
	OpIndex:          {"OpIndex", []int{}},
	OpCall:           {"OpCall", []int{1}}, // u8
	OpReturnValue:    {"OpReturnValue", []int{}},
	OpReturn:         {"OpReturn", []int{}},
	OpGetLocal:       {"OpGetLocal", []int{1}}, // u8 index
	OpSetLocal:       {"OpSetLocal", []int{1}},
	OpGetBuiltin:     {"OpGetBuiltin", []int{1}},
	OpClosure:        {"OpClosure", []int{2, 1}}, // index and free var
	OpGetFreeVar:     {"OpGetFreeVar", []int{1}},
	OpCurrnetClosure: {"OpCurrnetClosure", []int{}},
	OpLoop:           {"OpLoop", []int{2}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// 创建字节码指令
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1

	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}
	return instruction
}

// 将操作数大端编码转换到实际操作数
func ReadOperands(def *Definition, ins Instruction) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUnit16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUnit8(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUnit16(ins Instruction) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUnit8(ins Instruction) uint8 {
	return uint8(ins[0])
}
