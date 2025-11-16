package compiler

import (
	"fmt"

	"github.com/ninedraft/sulisp/interpreter/bytecode"
	"github.com/ninedraft/sulisp/language/ast"
)

type Compiler struct {
	builder       *builder
	scope         *Scope
	functions     map[string]*functionInfo
	functionOrder []string
	pendingCalls  []pendingCall
	err           error
}

func New() *Compiler {
	return &Compiler{}
}

func (compiler *Compiler) Compile(pkg *ast.Package) ([]bytecode.Command, error) {
	if pkg == nil {
		return nil, fmt.Errorf("package is nil")
	}

	compiler.builder = newBuilder()
	compiler.scope = nil
	compiler.functions = map[string]*functionInfo{}
	compiler.functionOrder = nil
	compiler.pendingCalls = nil
	compiler.err = nil

	var mainNodes []ast.Node
	for _, node := range pkg.Nodes {
		if fn, ok := node.(*ast.Function); ok {
			compiler.registerFunction(fn)
			continue
		}
		mainNodes = append(mainNodes, node)
	}

	for i, node := range mainNodes {
		compiler.compileNode(node)
		if compiler.err != nil {
			return nil, compiler.err
		}

		if i < len(mainNodes)-1 {
			compiler.builder.append(bytecode.Pop())
		}
	}

	compiler.builder.append(bytecode.Halt())
	compiler.compileFunctions()
	if compiler.err != nil {
		return nil, compiler.err
	}

	compiler.patchPendingCalls()
	if compiler.err != nil {
		return nil, compiler.err
	}

	return compiler.builder.Commands(), nil
}

func (compiler *Compiler) compileNode(node ast.Node) {
	if compiler.err != nil || node == nil {
		return
	}

	switch n := node.(type) {
	case *ast.Literal[int64]:
		compiler.builder.append(bytecode.Const(n.Value))
	case *ast.Literal[float64]:
		compiler.builder.append(bytecode.Const(n.Value))
	case *ast.Literal[string]:
		compiler.builder.append(bytecode.Const(n.Value))
	case *ast.Literal[bool]:
		compiler.builder.append(bytecode.Const(n.Value))
	case *ast.SpecialOp:
		compiler.compileSpecialOp(n)
	case *ast.If:
		compiler.compileIf(n)
	case *ast.Symbol:
		compiler.compileSymbol(n)
	case *ast.Assign:
		compiler.compileAssign(n)
	case *ast.SExp:
		compiler.compileSexpCall(n)
	default:
		compiler.err = fmt.Errorf("unsupported node %T", node)
	}
}

func (compiler *Compiler) registerFunction(fn *ast.Function) {
	if compiler.err != nil || fn == nil {
		return
	}

	if fn.Identifier == "" {
		compiler.err = fmt.Errorf("function name required")
		return
	}

	if _, ok := compiler.functions[fn.Identifier]; ok {
		compiler.err = fmt.Errorf("function %q already defined", fn.Identifier)
		return
	}

	compiler.functions[fn.Identifier] = &functionInfo{node: fn}
	compiler.functionOrder = append(compiler.functionOrder, fn.Identifier)
}

func (compiler *Compiler) compileSymbol(sym *ast.Symbol) {
	if sym == nil {
		return
	}

	if compiler.scope == nil {
		compiler.err = fmt.Errorf("symbol %s used outside of function scope", sym.Value)
		return
	}

	slot, ok := compiler.scope.resolve(sym.Value)
	if !ok {
		compiler.err = fmt.Errorf("symbol %s not defined", sym.Value)
		return
	}

	compiler.builder.append(bytecode.LoadLocal(slot))
}

func (compiler *Compiler) compileAssign(assign *ast.Assign) {
	if assign == nil {
		return
	}

	if compiler.scope == nil {
		compiler.err = fmt.Errorf("assignment outside of function scope")
		return
	}

	if assign.Target == nil {
		compiler.err = fmt.Errorf("assign target missing")
		return
	}

	compiler.compileNode(assign.Value)
	if compiler.err != nil {
		return
	}

	slot := compiler.scope.declare(assign.Target.Value)
	compiler.builder.append(bytecode.StoreLocal(slot))
	compiler.builder.append(bytecode.LoadLocal(slot))
}

func (compiler *Compiler) compileSexpCall(sexp *ast.SExp) {
	if sexp == nil {
		return
	}

	if len(sexp.Items) == 0 {
		compiler.err = fmt.Errorf("empty s-expression cannot be called")
		return
	}

	callee, ok := sexp.Items[0].(*ast.Symbol)
	if !ok {
		compiler.err = fmt.Errorf("s-expression head must be a symbol")
		return
	}

	args := sexp.Items[1:]
	for _, arg := range args {
		compiler.compileNode(arg)
		if compiler.err != nil {
			return
		}
	}

	index := compiler.builder.append(bytecode.Call(0, 0))

	compiler.pendingCalls = append(compiler.pendingCalls, pendingCall{
		index:    index,
		name:     callee.Value,
		argCount: len(args),
	})
}

func (compiler *Compiler) compileSpecialOp(op *ast.SpecialOp) {
	switch op.Op {
	case "+":
		compiler.compileVariadic(op.Items, bytecode.Add)
	case "*":
		compiler.compileVariadic(op.Items, bytecode.Mul)
	case "-":
		compiler.compileBinary(op.Items, bytecode.Sub)
	case "/":
		compiler.compileBinary(op.Items, bytecode.Div)
	default:
		compiler.err = fmt.Errorf("unsupported operator %s", op.Op)
	}
}

func (compiler *Compiler) compileVariadic(items []ast.Node, cmd bytecode.Command) {
	if len(items) == 0 {
		compiler.err = fmt.Errorf("operator %s requires at least one operand", cmd.Repr)
		return
	}

	compiler.compileNode(items[0])
	if compiler.err != nil {
		return
	}

	for _, operand := range items[1:] {
		compiler.compileNode(operand)
		if compiler.err != nil {
			return
		}
		compiler.builder.append(cmd)
	}
}

func (compiler *Compiler) compileBinary(items []ast.Node, cmd bytecode.Command) {
	if len(items) != 2 {
		compiler.err = fmt.Errorf("operator %s requires two operands", cmd.Repr)
		return
	}

	for _, operand := range items {
		compiler.compileNode(operand)
		if compiler.err != nil {
			return
		}
	}

	compiler.builder.append(cmd)
}

func (compiler *Compiler) compileIf(if_ *ast.If) {
	compiler.compileNode(if_.Cond)
	if compiler.err != nil {
		return
	}

	jumpFalse := compiler.builder.append(bytecode.JumpIfFalse(0))

	compiler.compileNode(if_.Then)
	if compiler.err != nil {
		return
	}

	jumpAfterThen := compiler.builder.append(bytecode.Jump(0))

	elseStart := compiler.builder.len()
	if if_.Else != nil {
		compiler.compileNode(if_.Else)
	} else {
		compiler.builder.append(bytecode.Null)
	}

	if compiler.err != nil {
		return
	}

	end := compiler.builder.len()

	compiler.builder.patch(jumpFalse, bytecode.JumpIfFalse(bytecode.PC(elseStart)))
	compiler.builder.patch(jumpAfterThen, bytecode.Jump(bytecode.PC(end)))
}

func (compiler *Compiler) compileFunctions() {
	prevScope := compiler.scope
	defer func() { compiler.scope = prevScope }()

	for _, name := range compiler.functionOrder {
		if compiler.err != nil {
			return
		}

		info := compiler.functions[name]
		info.entry = bytecode.PC(compiler.builder.len())

		compiler.scope = newScope(nil)
		for _, param := range info.node.Parameters {
			compiler.scope.declare(param.Value)
		}

		compiler.compileNode(info.node.Body)
		if compiler.err != nil {
			return
		}

		compiler.builder.append(bytecode.Return)
	}
}

func (compiler *Compiler) patchPendingCalls() {
	for _, pending := range compiler.pendingCalls {
		info, ok := compiler.functions[pending.name]
		if !ok {
			compiler.err = fmt.Errorf("undefined function %s", pending.name)
			return
		}
		if pending.argCount != len(info.node.Parameters) {
			compiler.err = fmt.Errorf("function %s called with wrong number of args", pending.name)
			return
		}

		compiler.builder.patch(pending.index, bytecode.Call(info.entry, len(info.node.Parameters)))
	}
}

type functionInfo struct {
	node  *ast.Function
	entry bytecode.PC
}

type pendingCall struct {
	index    int
	name     string
	argCount int
}

type Scope struct {
	parent *Scope
	slots  map[string]int
	next   int
}

func newScope(parent *Scope) *Scope {
	return &Scope{
		parent: parent,
		slots:  map[string]int{},
	}
}

func (scope *Scope) declare(name string) int {
	if slot, ok := scope.slots[name]; ok {
		return slot
	}

	slot := scope.next
	scope.slots[name] = slot
	scope.next++
	return slot
}

func (scope *Scope) resolve(name string) (int, bool) {
	if scope == nil {
		return 0, false
	}

	if slot, ok := scope.slots[name]; ok {
		return slot, true
	}
	if scope.parent != nil {
		return scope.parent.resolve(name)
	}
	return 0, false
}

type builder struct {
	commands []bytecode.Command
}

func newBuilder() *builder {
	return &builder{}
}

func (builder *builder) append(cmd bytecode.Command) int {
	builder.commands = append(builder.commands, cmd)
	return len(builder.commands) - 1
}

func (builder *builder) patch(index int, cmd bytecode.Command) {
	builder.commands[index] = cmd
}

func (builder *builder) len() int {
	return len(builder.commands)
}

func (builder *builder) Commands() []bytecode.Command {
	cpy := make([]bytecode.Command, len(builder.commands))
	copy(cpy, builder.commands)
	return cpy
}
