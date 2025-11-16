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
	effectStack   []effectContext
	currentFunc   *functionInfo
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
	case *ast.Keyword:
		compiler.builder.append(bytecode.Const(n.Value))
	case *ast.SpecialOp:
		compiler.compileSpecialOp(n)
	case *ast.If:
		compiler.compileIf(n)
	case *ast.Symbol:
		compiler.compileSymbol(n)
	case *ast.Assign:
		compiler.compileAssign(n)
	case *ast.Sequence:
		compiler.compileSequence(n)
	case *ast.While:
		compiler.compileWhile(n)
	case *ast.Let:
		compiler.compileLet(n)
	case *ast.Handle:
		compiler.compileHandle(n)
	case *ast.SExp:
		compiler.compileSexpCall(n)
	case *ast.Match:
		compiler.compileMatch(n)
	default:
		compiler.err = fmt.Errorf("unsupported node %T", node)
	}
}

func (compiler *Compiler) compileSequence(seq *ast.Sequence) {
	for i, item := range seq.Items {
		compiler.compileNode(item)
		if compiler.err != nil {
			return
		}

		if i < len(seq.Items)-1 {
			compiler.builder.append(bytecode.Pop())
		}
	}
}

func (compiler *Compiler) compileWhile(loop *ast.While) {
	if loop == nil {
		return
	}

	start := compiler.builder.len()

	compiler.compileNode(loop.Cond)
	if compiler.err != nil {
		return
	}

	exitJump := compiler.builder.append(bytecode.JumpIfFalse(0))

	compiler.compileNode(loop.Body)
	if compiler.err != nil {
		return
	}

	compiler.builder.append(bytecode.Pop())
	compiler.builder.append(bytecode.Jump(bytecode.PC(start)))

	end := compiler.builder.len()
	compiler.builder.patch(exitJump, bytecode.JumpIfFalse(bytecode.PC(end)))
	compiler.builder.append(bytecode.Null)
}

func (compiler *Compiler) compileLet(let *ast.Let) {
	if let == nil {
		return
	}

	if compiler.scope == nil {
		compiler.err = fmt.Errorf("let outside of function scope")
		return
	}

	prev := compiler.scope
	compiler.scope = compiler.scope.child()
	defer func() { compiler.scope = prev }()

	for _, binding := range let.Bindings {
		compiler.compileNode(binding.Value)
		if compiler.err != nil {
			return
		}

		slot := compiler.scope.declareFresh(binding.Identifier)
		compiler.builder.append(bytecode.StoreLocal(slot))
	}

	for i, item := range let.Body {
		compiler.compileNode(item)
		if compiler.err != nil {
			return
		}

		if i < len(let.Body)-1 {
			compiler.builder.append(bytecode.Pop())
		}
	}
}

func (compiler *Compiler) compileHandle(handle *ast.Handle) {
	if handle == nil {
		return
	}

	operations := make(map[string]*functionInfo, len(handle.Operations))

	for _, op := range handle.Operations {
		fn, ok := op.Body.(*ast.Function)
		if !ok {
			compiler.err = fmt.Errorf("handle operation %s must be a function", op.OpName)
			return
		}

		name := fmt.Sprintf("handle:%s:%s:%d", handle.Effect, op.OpName, len(compiler.functionOrder))
		info := compiler.registerFunctionWithName(name, fn, true, handle.Effect, op.OpName)
		if compiler.err != nil {
			return
		}

		operations[op.OpName] = info
	}

	ctx := effectContext{
		name:       handle.Effect,
		operations: operations,
	}

	compiler.effectStack = append(compiler.effectStack, ctx)
	defer func() {
		compiler.effectStack = compiler.effectStack[:len(compiler.effectStack)-1]
	}()

	for i, item := range handle.Body {
		compiler.compileNode(item)
		if compiler.err != nil {
			return
		}

		if i < len(handle.Body)-1 {
			compiler.builder.append(bytecode.Pop())
		}
	}
}

func (compiler *Compiler) registerFunction(fn *ast.Function) {
	if compiler.err != nil || fn == nil {
		return
	}

	compiler.registerFunctionWithName(fn.Identifier, fn, false, "", "")
}

func (compiler *Compiler) registerFunctionWithName(name string, fn *ast.Function, requiresContinuation bool, effectName, opName string) *functionInfo {
	if compiler.err != nil || fn == nil {
		return nil
	}

	if name == "" {
		compiler.err = fmt.Errorf("function name required")
		return nil
	}

	if _, ok := compiler.functions[name]; ok {
		compiler.err = fmt.Errorf("function %q already defined", name)
		return nil
	}

	info := &functionInfo{
		node:                 fn,
		entry:                0,
		params:               len(fn.Parameters),
		effectName:           effectName,
		operationName:        opName,
		requiresContinuation: requiresContinuation,
		name:                 name,
	}

	if requiresContinuation {
		if info.params == 0 {
			compiler.err = fmt.Errorf("effect %s.%s must have continuation parameter", effectName, opName)
			return nil
		}
		info.continuationParam = fn.Parameters[info.params-1].Value
		info.continuationSlot = info.params - 1
	}

	compiler.functions[name] = info
	compiler.functionOrder = append(compiler.functionOrder, name)

	return info
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

	if compiler.compileContinuationCall(sexp) {
		return
	}

	if compiler.compileEffectCall(sexp) {
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

func (compiler *Compiler) compileContinuationCall(sexp *ast.SExp) bool {
	info := compiler.currentFunc
	if info == nil || !info.requiresContinuation {
		return false
	}

	head, ok := sexp.Items[0].(*ast.Symbol)
	if !ok {
		return false
	}

	if head.Value != info.continuationParam {
		return false
	}

	args := sexp.Items[1:]
	for _, arg := range args {
		compiler.compileNode(arg)
		if compiler.err != nil {
			return true
		}
	}

	compiler.builder.append(bytecode.LoadLocal(info.continuationSlot))
	compiler.builder.append(bytecode.InvokeContinuation(len(args)))

	return true
}

func (compiler *Compiler) compileEffectCall(sexp *ast.SExp) bool {
	if len(sexp.Items) < 2 {
		return false
	}

	head, ok := sexp.Items[0].(*ast.Symbol)
	if !ok {
		return false
	}

	op, ok := sexp.Items[1].(*ast.Symbol)
	if !ok {
		return false
	}

	info := compiler.lookupEffectOperation(head.Value, op.Value)
	if info == nil {
		return false
	}

	args := sexp.Items[2:]
	expected := info.params - 1
	if expected < 0 {
		expected = 0
	}
	if len(args) != expected {
		compiler.err = fmt.Errorf("effect %s.%s expects %d args, got %d", head.Value, op.Value, expected, len(args))
		return true
	}

	for _, arg := range args {
		compiler.compileNode(arg)
		if compiler.err != nil {
			return true
		}
	}

	compiler.builder.append(bytecode.PushContinuation())
	index := compiler.builder.append(bytecode.Call(0, info.params))

	compiler.pendingCalls = append(compiler.pendingCalls, pendingCall{
		index:    index,
		name:     info.name,
		argCount: info.params,
	})

	return true
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

func (compiler *Compiler) compileMatch(match *ast.Match) {
	if match == nil {
		return
	}

	if compiler.scope == nil {
		compiler.err = fmt.Errorf("match outside of function scope")
		return
	}

	if len(match.Cases) == 0 {
		compiler.err = fmt.Errorf("match requires at least one case")
		return
	}

	compiler.compileNode(match.Expr)
	if compiler.err != nil {
		return
	}

	targetSlot := compiler.scope.declareFresh("__match_target")
	compiler.builder.append(bytecode.StoreLocal(targetSlot))

	var jumpAfterIdxs []int

	for _, mcase := range match.Cases {
		if len(mcase.Body) == 0 {
			compiler.err = fmt.Errorf("match case body missing")
			return
		}

		jumpAfter := compiler.compileMatchCase(mcase, targetSlot)
		if compiler.err != nil {
			return
		}
		jumpAfterIdxs = append(jumpAfterIdxs, jumpAfter)
	}

	compiler.builder.append(bytecode.Null)
	finalPC := compiler.builder.len()
	for _, idx := range jumpAfterIdxs {
		compiler.builder.patch(idx, bytecode.Jump(bytecode.PC(finalPC)))
	}
}

func (compiler *Compiler) compileMatchCase(mcase *ast.MatchCase, targetSlot int) int {
	if compiler.scope == nil {
		compiler.err = fmt.Errorf("match case requires function scope")
		return 0
	}

	var bindingName string

	switch pattern := mcase.Pattern.(type) {
	case *ast.PatternLiteral:
		compiler.builder.append(bytecode.LoadLocal(targetSlot))
		compiler.compileNode(pattern.Value)
		if compiler.err != nil {
			return 0
		}
		compiler.builder.append(bytecode.Equal)
	case *ast.PatternWildcard:
		compiler.builder.append(bytecode.Const(true))
	case *ast.PatternVariable:
		compiler.builder.append(bytecode.Const(true))
		bindingName = pattern.Identifier
	default:
		compiler.err = fmt.Errorf("unsupported match pattern %T", mcase.Pattern)
		return 0
	}

	jumpFalse := compiler.builder.append(bytecode.JumpIfFalse(0))

	prevScope := compiler.scope
	caseScope := compiler.scope.child()
	compiler.scope = caseScope

	if bindingName != "" {
		compiler.builder.append(bytecode.LoadLocal(targetSlot))
		slot := compiler.scope.declare(bindingName)
		compiler.builder.append(bytecode.StoreLocal(slot))
	}

	for i, node := range mcase.Body {
		compiler.compileNode(node)
		if compiler.err != nil {
			compiler.scope = prevScope
			return 0
		}
		if i < len(mcase.Body)-1 {
			compiler.builder.append(bytecode.Pop())
		}
	}

	jumpAfter := compiler.builder.append(bytecode.Jump(0))

	caseBodyEnd := compiler.builder.len()
	compiler.builder.patch(jumpFalse, bytecode.JumpIfFalse(bytecode.PC(caseBodyEnd)))

	return jumpAfter
}

func (compiler *Compiler) compileFunctions() {
	prevScope := compiler.scope
	defer func() { compiler.scope = prevScope }()

	for i := 0; i < len(compiler.functionOrder); i++ {
		name := compiler.functionOrder[i]
		if compiler.err != nil {
			return
		}

		info := compiler.functions[name]
		info.entry = bytecode.PC(compiler.builder.len())

		compiler.scope = newScope(nil)
		for _, param := range info.node.Parameters {
			compiler.scope.declare(param.Value)
		}
		prevFunc := compiler.currentFunc
		compiler.currentFunc = info

		compiler.compileNode(info.node.Body)

		compiler.currentFunc = prevFunc
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
	node                 *ast.Function
	entry                bytecode.PC
	params               int
	requiresContinuation bool
	effectName           string
	operationName        string
	continuationParam    string
	continuationSlot     int
	name                 string
}

type pendingCall struct {
	index    int
	name     string
	argCount int
}

type effectContext struct {
	name       string
	operations map[string]*functionInfo
}

func (compiler *Compiler) lookupEffectOperation(effect, op string) *functionInfo {
	for i := len(compiler.effectStack) - 1; i >= 0; i-- {
		ctx := compiler.effectStack[i]
		if ctx.name != effect {
			continue
		}

		info, ok := ctx.operations[op]
		if ok {
			return info
		}
	}

	return nil
}

type Scope struct {
	parent  *Scope
	slots   map[string]int
	counter *int
}

func newScope(parent *Scope) *Scope {
	if parent == nil {
		counter := 0
		return &Scope{
			parent:  nil,
			slots:   map[string]int{},
			counter: &counter,
		}
	}

	return &Scope{
		parent:  parent,
		slots:   map[string]int{},
		counter: parent.counter,
	}
}

func (scope *Scope) declare(name string) int {
	if slot, ok := scope.slots[name]; ok {
		return slot
	}

	slot := scope.newSlot()
	scope.slots[name] = slot
	return slot
}

func (scope *Scope) declareFresh(name string) int {
	slot := scope.newSlot()
	scope.slots[name] = slot
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

func (scope *Scope) child() *Scope {
	return &Scope{
		parent:  scope,
		slots:   map[string]int{},
		counter: scope.counter,
	}
}

func (scope *Scope) newSlot() int {
	slot := *scope.counter
	*scope.counter++
	return slot
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
