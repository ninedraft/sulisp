// Package compiler translates generic Sulisp syntax into register VM programs.
package compiler

import (
	"fmt"

	"github.com/ninedraft/sulisp/language/ast"
	"github.com/ninedraft/sulisp/language/tokens"
	"github.com/ninedraft/sulisp/vm"
)

// Error is a compilation failure at a source position.
type Error struct {
	Pos tokens.Position
	Msg string
}

func (compileError *Error) Error() string {
	return fmt.Sprintf("%s: %s", compileError.Pos, compileError.Msg)
}

// Compile translates top-level expressions into an executable Program.
// Registers are reused between top-level expressions because no result
// retention or bindings are part of this minimal compiler.
func Compile(nodes []ast.Node) (vm.Program, error) {
	compiler := &compiler{}
	for _, node := range nodes {
		compiler.next = 0
		destination := compiler.alloc()
		if err := compiler.expr(node, destination); err != nil {
			return vm.Program{}, err
		}
	}
	return vm.Program{Ops: compiler.ops, NumRegs: compiler.max}, nil
}

type compiler struct {
	ops  vm.Tape
	next vm.Reg
	max  int
}

func (compiler *compiler) alloc() vm.Reg {
	register := compiler.next
	compiler.next++
	compiler.max = max(compiler.max, int(compiler.next))
	return register
}

func (compiler *compiler) free(register vm.Reg) {
	compiler.next = register
}

func (compiler *compiler) emit(operation vm.Operation) {
	compiler.ops = append(compiler.ops, operation)
}

func (compiler *compiler) expr(node ast.Node, destination vm.Reg) error {
	switch node := node.(type) {
	case ast.Atom[int64]:
		compiler.emit(vm.LoadConst{Pos: node.Pos().From, Dst: destination, Val: vm.Int(node.Value)})
		return nil
	case *ast.Atom[int64]:
		if node == nil {
			return unsupportedNil()
		}
		return compiler.expr(*node, destination)
	case ast.Atom[float64]:
		compiler.emit(vm.LoadConst{Pos: node.Pos().From, Dst: destination, Val: vm.Float(node.Value)})
		return nil
	case *ast.Atom[float64]:
		if node == nil {
			return unsupportedNil()
		}
		return compiler.expr(*node, destination)
	case ast.Atom[bool]:
		compiler.emit(vm.LoadConst{Pos: node.Pos().From, Dst: destination, Val: vm.Bool(node.Value)})
		return nil
	case *ast.Atom[bool]:
		if node == nil {
			return unsupportedNil()
		}
		return compiler.expr(*node, destination)
	case ast.Atom[string]:
		return compiler.stringAtom(node, destination)
	case *ast.Atom[string]:
		if node == nil {
			return unsupportedNil()
		}
		return compiler.stringAtom(*node, destination)
	case ast.List:
		return compiler.form(node, destination)
	case *ast.List:
		if node == nil {
			return unsupportedNil()
		}
		return compiler.form(*node, destination)
	default:
		return &Error{Pos: node.Pos().From, Msg: fmt.Sprintf("unsupported node %T", node)}
	}
}

func unsupportedNil() error {
	return &Error{Msg: "unsupported nil node"}
}

func (compiler *compiler) stringAtom(atom ast.Atom[string], destination vm.Reg) error {
	position := atom.Pos().From
	switch atom.Kind {
	case tokens.TokenStr:
		compiler.emit(vm.LoadConst{Pos: position, Dst: destination, Val: vm.String(atom.Value)})
		return nil
	case tokens.TokenKeyword:
		compiler.emit(vm.LoadConst{Pos: position, Dst: destination, Val: vm.Keyword(atom.Value)})
		return nil
	case tokens.TokenSymbol:
		return &Error{Pos: position, Msg: fmt.Sprintf("unknown symbol %q", atom.Value)}
	default:
		return &Error{Pos: position, Msg: fmt.Sprintf("unsupported string atom kind %s", atom.Kind)}
	}
}

func (compiler *compiler) form(list ast.List, destination vm.Reg) error {
	if len(list.Items) == 0 {
		return &Error{Pos: list.Pos().From, Msg: "empty form"}
	}

	head, ok := symbol(list.Items[0])
	if !ok {
		return &Error{
			Pos: list.Items[0].Pos().From,
			Msg: fmt.Sprintf("form head must be a symbol, got %s", syntaxKind(list.Items[0])),
		}
	}
	arguments := list.Items[1:]
	switch head.Value {
	case "print":
		return compiler.print(head, arguments, destination)
	case "+", "-", "*", "/":
		return compiler.arithmetic(head, arguments, destination)
	case "=", "<", ">", "<=", ">=":
		return compiler.comparison(head, arguments, destination)
	case "if":
		return compiler.ifForm(head, arguments, destination)
	case "not":
		return compiler.not(head, arguments, destination)
	case "and", "or":
		return compiler.logic(head, arguments, destination)
	default:
		return &Error{Pos: head.Pos().From, Msg: fmt.Sprintf("unknown form %q", head.Value)}
	}
}

func symbol(node ast.Node) (ast.Atom[string], bool) {
	switch node := node.(type) {
	case ast.Atom[string]:
		return node, node.Kind == tokens.TokenSymbol
	case *ast.Atom[string]:
		if node != nil {
			return *node, node.Kind == tokens.TokenSymbol
		}
	}
	return ast.Atom[string]{}, false
}

func syntaxKind(node ast.Node) string {
	switch node := node.(type) {
	case ast.Atom[int64], *ast.Atom[int64]:
		return "integer"
	case ast.Atom[float64], *ast.Atom[float64]:
		return "float"
	case ast.Atom[bool], *ast.Atom[bool]:
		return "boolean"
	case ast.Atom[string]:
		return stringSyntaxKind(node.Kind)
	case *ast.Atom[string]:
		if node == nil {
			return "nil"
		}
		return stringSyntaxKind(node.Kind)
	case ast.List, *ast.List:
		return "list"
	default:
		return fmt.Sprintf("%T", node)
	}
}

func stringSyntaxKind(kind tokens.TokenKind) string {
	switch kind {
	case tokens.TokenStr:
		return "string"
	case tokens.TokenKeyword:
		return "keyword"
	case tokens.TokenSymbol:
		return "symbol"
	default:
		return kind.String()
	}
}

func (compiler *compiler) print(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) != 1 {
		return arityError(head, "1 argument", len(arguments))
	}
	if err := compiler.expr(arguments[0], destination); err != nil {
		return err
	}
	compiler.emit(vm.Print{Pos: head.Pos().From, Src: destination})
	return nil
}

func (compiler *compiler) arithmetic(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) < 2 {
		return arityError(head, "at least 2 arguments", len(arguments))
	}
	if err := compiler.expr(arguments[0], destination); err != nil {
		return err
	}

	temporary := compiler.alloc()
	defer compiler.free(temporary)
	for _, argument := range arguments[1:] {
		if err := compiler.expr(argument, temporary); err != nil {
			return err
		}
		position := head.Pos().From
		switch head.Value {
		case "+":
			compiler.emit(vm.Add{Pos: position, Dst: destination, A: destination, B: temporary})
		case "-":
			compiler.emit(vm.Sub{Pos: position, Dst: destination, A: destination, B: temporary})
		case "*":
			compiler.emit(vm.Mul{Pos: position, Dst: destination, A: destination, B: temporary})
		case "/":
			compiler.emit(vm.Div{Pos: position, Dst: destination, A: destination, B: temporary})
		}
	}
	return nil
}

func (compiler *compiler) comparison(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) != 2 {
		return arityError(head, "2 arguments", len(arguments))
	}
	if err := compiler.expr(arguments[0], destination); err != nil {
		return err
	}

	temporary := compiler.alloc()
	defer compiler.free(temporary)
	if err := compiler.expr(arguments[1], temporary); err != nil {
		return err
	}
	position := head.Pos().From
	switch head.Value {
	case "=":
		compiler.emit(vm.Eq{Pos: position, Dst: destination, A: destination, B: temporary})
	case "<":
		compiler.emit(vm.Lt{Pos: position, Dst: destination, A: destination, B: temporary})
	case ">":
		compiler.emit(vm.Gt{Pos: position, Dst: destination, A: destination, B: temporary})
	case "<=":
		compiler.emit(vm.Le{Pos: position, Dst: destination, A: destination, B: temporary})
	case ">=":
		compiler.emit(vm.Ge{Pos: position, Dst: destination, A: destination, B: temporary})
	}
	return nil
}

func (compiler *compiler) ifForm(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) != 3 {
		return arityError(head, "3 arguments (cond then else)", len(arguments))
	}
	condition, thenBranch, elseBranch := arguments[0], arguments[1], arguments[2]

	conditionRegister := compiler.alloc()
	if err := compiler.expr(condition, conditionRegister); err != nil {
		return err
	}
	elseJump := len(compiler.ops)
	compiler.emit(vm.JumpIfFalse{
		Pos:    condition.Pos().From,
		Cond:   conditionRegister,
		Target: -1,
	})
	compiler.free(conditionRegister)

	if err := compiler.expr(thenBranch, destination); err != nil {
		return err
	}
	endJump := len(compiler.ops)
	compiler.emit(vm.Jump{Pos: head.Pos().From, Target: -1})

	compiler.patchJumpIfFalse(elseJump, len(compiler.ops))
	if err := compiler.expr(elseBranch, destination); err != nil {
		return err
	}
	compiler.patchJump(endJump, len(compiler.ops))
	return nil
}

func (compiler *compiler) patchJump(index, target int) {
	operation := compiler.ops[index].(vm.Jump)
	operation.Target = target
	compiler.ops[index] = operation
}

func (compiler *compiler) patchJumpIfFalse(index, target int) {
	operation := compiler.ops[index].(vm.JumpIfFalse)
	operation.Target = target
	compiler.ops[index] = operation
}

func (compiler *compiler) not(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) != 1 {
		return arityError(head, "1 argument", len(arguments))
	}
	if err := compiler.expr(arguments[0], destination); err != nil {
		return err
	}
	compiler.emit(vm.Not{Pos: head.Pos().From, Dst: destination, Src: destination})
	return nil
}

func (compiler *compiler) logic(head ast.Atom[string], arguments []ast.Node, destination vm.Reg) error {
	if len(arguments) < 2 {
		return arityError(head, "at least 2 arguments", len(arguments))
	}
	return compiler.ifForm(head, desugarLogic(head, arguments), destination)
}

func desugarLogic(head ast.Atom[string], arguments []ast.Node) []ast.Node {
	rest := arguments[1:]
	var restExpression ast.Node = rest[0]
	if len(rest) > 1 {
		items := make([]ast.Node, 1, len(rest)+1)
		items[0] = head
		items = append(items, rest...)
		restExpression = ast.List{PosRange: head.PosRange, Items: items}
	}
	boolean := func(value bool) ast.Node {
		return ast.Atom[bool]{PosRange: head.PosRange, Value: value}
	}
	if head.Value == "and" {
		return []ast.Node{arguments[0], restExpression, boolean(false)}
	}
	return []ast.Node{arguments[0], boolean(true), restExpression}
}

func arityError(head ast.Atom[string], expected string, got int) error {
	return &Error{
		Pos: head.Pos().From,
		Msg: fmt.Sprintf("%s expects %s, got %d", head.Value, expected, got),
	}
}
