package ast

import (
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/ninedraft/sulisp/language/tokens"
)

type Node interface {
	fmt.Stringer
	Equal(other Node) bool
	Name() string
	Pos() PosRange
	Clone() Node
}

type PosRange struct {
	From, To tokens.Position
}

func (pos PosRange) Pos() PosRange {
	return pos
}

type LiteralValue interface {
	string | int64 | float64 | bool
}

type Literal[L LiteralValue] struct {
	PosRange
	Value L
}

func (lit *Literal[E]) Equal(other Node) bool {
	if lit == nil {
		return other == nil
	}

	o, ok := other.(*Literal[E])
	if !ok {
		return false
	}

	return lit.Value == o.Value
}

func (*Literal[L]) Name() string {
	var v L
	switch any(v).(type) {
	case string:
		return "string"
	case int64:
		return "int"
	case float64:
		return "float"
	case bool:
		return "bool"
	}

	return fmt.Sprintf("literal[%T]", v)
}

func (lit *Literal[L]) String() string {
	return fmt.Sprint(lit.Value)
}

func (lit *Literal[L]) Clone() Node {
	return shallow(lit)
}

type Package struct {
	PosRange
	Nodes []Node
}

func (pkg *Package) Equal(other Node) bool {
	if pkg == nil {
		return other == nil
	}

	if o, _ := other.(*Package); o != nil {
		return equalSlices(pkg.Nodes, o.Nodes)
	}

	return false
}

func (*Package) Name() string {
	return "package"
}

func (pkg *Package) String() string {
	str := &strings.Builder{}
	joinStringers(str, "\n\n", pkg.Nodes)
	return str.String()
}

func (pkg *Package) Clone() Node {
	if pkg == nil {
		return nil
	}

	clone := *pkg
	clone.Nodes = cloneSlice(pkg.Nodes)

	return &clone
}

type Symbol struct {
	PosRange
	Value string
}

func (sym *Symbol) Equal(node Node) bool {
	if sym == nil {
		return node == nil
	}

	if o, _ := node.(*Symbol); o != nil {
		return sym.Value == o.Value
	}

	return false
}

func (*Symbol) Name() string { return tokens.TokenSymbol.String() }

func (sym *Symbol) String() string { return sym.Value }

func (sym *Symbol) Clone() Node {
	return shallow(sym)
}

type Keyword struct {
	PosRange
	Value string
}

func (kw *Keyword) Equal(node Node) bool {
	if node == nil {
		return node == nil
	}

	if o, _ := node.(*Keyword); o != nil {
		return kw.Value == o.Value
	}

	return false
}

func (*Keyword) Name() string { return tokens.TokenKeyword.String() }

func (kw *Keyword) String() string { return kw.Value }

func (kw *Keyword) Clone() Node {
	return shallow(kw)
}

type SExp struct {
	PosRange
	Items []Node
}

func NewSexp(items ...Node) *SExp {
	return &SExp{
		Items: items,
	}
}

func (sexp *SExp) Name() string {
	return "s-expr"
}

func (sexp *SExp) String() string {
	str := &strings.Builder{}

	str.WriteRune('(')
	joinStringers(str, " ", sexp.Items)
	str.WriteRune(')')

	return str.String()
}

func (sexp *SExp) Equal(other Node) bool {
	if sexp == nil {
		return other == nil
	}

	if o, _ := other.(*SExp); o != nil {
		return equalSlices(sexp.Items, o.Items)
	}

	return false
}

func (sexp *SExp) Clone() Node {
	if sexp == nil {
		return nil
	}

	clone := *sexp
	clone.Items = cloneSlice(sexp.Items)

	return &clone
}

type DotSelector struct {
	PosRange
	Left, Right Node // left.right
}

func (dot *DotSelector) Equal(other Node) bool {
	if dot == nil {
		return other == nil
	}

	if o, _ := other.(*DotSelector); o != nil {
		return dot.Left.Equal(o.Left) && dot.Right.Equal(o.Right)
	}

	return false
}

func (*DotSelector) Name() string {
	return "dot-selector"
}

func (dot *DotSelector) String() string {
	return "(. " + dot.Left.String() + " " + dot.Right.String() + ")"
}

func (dot *DotSelector) Clone() Node {
	if dot == nil {
		return nil
	}

	clone := shallow(dot)
	clone.Left = Clone(dot.Left)
	clone.Right = Clone(dot.Right)

	return clone
}

type Sequence struct {
	PosRange
	Items []Node
}

func (seq *Sequence) Name() string {
	return "sequence"
}

func (seq *Sequence) String() string {
	if seq == nil {
		return "nil-sequence"
	}

	str := &strings.Builder{}
	str.WriteString("(begin")
	for _, item := range seq.Items {
		str.WriteRune(' ')
		str.WriteString(item.String())
	}
	str.WriteString(")")
	return str.String()
}

func (seq *Sequence) Equal(other Node) bool {
	if seq == nil {
		return other == nil
	}

	o, ok := other.(*Sequence)
	if !ok {
		return false
	}

	return equalSlices(seq.Items, o.Items)
}

func (seq *Sequence) Clone() Node {
	if seq == nil {
		return nil
	}

	clone := *seq
	clone.Items = cloneSlice(seq.Items)
	return &clone
}

type Binding struct {
	PosRange
	Identifier string
	Value      Node
}

func (binding *Binding) Name() string { return "binding" }

func (binding *Binding) Equal(other Node) bool {
	if binding == nil {
		return other == nil
	}

	o, ok := other.(*Binding)
	if !ok {
		return false
	}

	return binding.Identifier == o.Identifier && equalNodes(binding.Value, o.Value)
}

func (binding *Binding) Clone() Node {
	if binding == nil {
		return nil
	}

	clone := *binding
	clone.Value = Clone(binding.Value)
	return &clone
}

func (binding *Binding) String() string {
	if binding == nil {
		return "nil-binding"
	}

	return "(" + binding.Identifier + " " + binding.Value.String() + ")"
}

type Let struct {
	PosRange
	Bindings []*Binding
	Body     []Node
}

func (let *Let) Name() string {
	return "let"
}

func (let *Let) String() string {
	if let == nil {
		return "nil-let"
	}

	str := &strings.Builder{}
	str.WriteString("(let (")
	for i, binding := range let.Bindings {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteString(binding.String())
	}
	str.WriteString(")")
	for _, item := range let.Body {
		str.WriteRune(' ')
		str.WriteString(item.String())
	}
	str.WriteString(")")
	return str.String()
}

func (let *Let) Equal(other Node) bool {
	if let == nil {
		return other == nil
	}

	o, ok := other.(*Let)
	if !ok {
		return false
	}

	return equalSlices(let.Bindings, o.Bindings) &&
		equalSlices(let.Body, o.Body)
}

func (let *Let) Clone() Node {
	if let == nil {
		return nil
	}

	clone := *let
	clone.Bindings = cloneSlice(let.Bindings)
	clone.Body = cloneSlice(let.Body)
	return &clone
}

type Handle struct {
	PosRange
	Effect     string
	Operations []*HandleOp
	Body       []Node
}

func (handle *Handle) Name() string { return "handle" }

func (handle *Handle) Equal(other Node) bool {
	if handle == nil {
		return other == nil
	}
	o, ok := other.(*Handle)
	if !ok {
		return false
	}
	return handle.Effect == o.Effect &&
		equalSlices(handle.Operations, o.Operations) &&
		equalSlices(handle.Body, o.Body)
}

func (handle *Handle) Clone() Node {
	if handle == nil {
		return nil
	}
	clone := *handle
	clone.Operations = cloneSlice(handle.Operations)
	clone.Body = cloneSlice(handle.Body)
	return &clone
}

func (handle *Handle) String() string {
	if handle == nil {
		return "nil-handle"
	}
	str := &strings.Builder{}
	str.WriteString("(handle ")
	str.WriteString(handle.Effect)
	str.WriteString(" (")
	for i, op := range handle.Operations {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteString(op.String())
	}
	str.WriteString(")")
	for _, item := range handle.Body {
		str.WriteRune(' ')
		str.WriteString(item.String())
	}
	str.WriteString(")")
	return str.String()
}

type HandleOp struct {
	PosRange
	OpName string
	Body   Node
}

func (op *HandleOp) Name() string { return "handle-op" }

func (op *HandleOp) Equal(other Node) bool {
	if op == nil {
		return other == nil
	}
	o, ok := other.(*HandleOp)
	if !ok {
		return false
	}
	return op.OpName == o.OpName && equalNodes(op.Body, o.Body)
}

func (op *HandleOp) Clone() Node {
	if op == nil {
		return nil
	}
	clone := *op
	clone.Body = Clone(op.Body)
	return &clone
}

func (op *HandleOp) String() string {
	if op == nil {
		return "nil-handle-op"
	}
	return "(" + op.OpName + " " + op.Body.String() + ")"
}

type While struct {
	PosRange
	Cond Node
	Body Node
}

func (while *While) Name() string {
	return "while"
}

func (while *While) String() string {
	if while == nil {
		return "nil-while"
	}

	str := &strings.Builder{}
	str.WriteString("(while ")
	if while.Cond != nil {
		str.WriteString(while.Cond.String())
	}
	str.WriteString(" ")
	if while.Body != nil {
		str.WriteString(while.Body.String())
	}
	str.WriteString(")")
	return str.String()
}

func (while *While) Equal(other Node) bool {
	if while == nil {
		return other == nil
	}

	o, ok := other.(*While)
	if !ok {
		return false
	}

	return equalNodes(while.Cond, o.Cond) && equalNodes(while.Body, o.Body)
}

func (while *While) Clone() Node {
	if while == nil {
		return nil
	}

	clone := *while
	clone.Cond = Clone(while.Cond)
	clone.Body = Clone(while.Body)
	return &clone
}

type Function struct {
	PosRange
	Identifier string
	Parameters []*Symbol
	Body       Node
}

func (fn *Function) Name() string {
	return "function"
}

func (fn *Function) Equal(other Node) bool {
	if fn == nil {
		return other == nil
	}

	o, ok := other.(*Function)
	if !ok {
		return false
	}

	return fn.Identifier == o.Identifier &&
		equalSlices(fn.Parameters, o.Parameters) &&
		equalNodes(fn.Body, o.Body)
}

func (fn *Function) Clone() Node {
	if fn == nil {
		return nil
	}

	clone := *fn
	clone.Parameters = cloneSlice(fn.Parameters)
	clone.Body = Clone(fn.Body)
	return &clone
}

func (fn *Function) String() string {
	if fn == nil {
		return "nil-function"
	}

	str := &strings.Builder{}
	str.WriteString("(fn ")
	str.WriteString(fn.Identifier)
	str.WriteString(" (")
	for i, param := range fn.Parameters {
		if i > 0 {
			str.WriteString(" ")
		}
		str.WriteString(param.String())
	}
	str.WriteString(") ")
	if fn.Body != nil {
		str.WriteString(fn.Body.String())
	}
	str.WriteString(")")
	return str.String()
}

type Assign struct {
	PosRange
	Target *Symbol
	Value  Node
}

func (assign *Assign) Name() string {
	return "assign"
}

func (assign *Assign) Equal(other Node) bool {
	if assign == nil {
		return other == nil
	}

	o, ok := other.(*Assign)
	if !ok {
		return false
	}

	return equalNodes(assign.Target, o.Target) &&
		equalNodes(assign.Value, o.Value)
}

func (assign *Assign) Clone() Node {
	if assign == nil {
		return nil
	}

	clone := *assign
	if assign.Target != nil {
		clone.Target = Clone(assign.Target)
	}
	clone.Value = Clone(assign.Value)
	return &clone
}

func (assign *Assign) String() string {
	if assign == nil {
		return "nil-assign"
	}

	str := &strings.Builder{}
	str.WriteString("(assign ")
	if assign.Target != nil {
		str.WriteString(assign.Target.String())
	}
	str.WriteString(" ")
	if assign.Value != nil {
		str.WriteString(assign.Value.String())
	}
	str.WriteString(")")
	return str.String()
}

// Pattern represents a match pattern part.
type Pattern interface {
	patternNode()
	Equal(Pattern) bool
	Clone() Pattern
	String() string
}

type PatternLiteral struct {
	PosRange
	Value Node
}

func (*PatternLiteral) patternNode() {}

func (lit *PatternLiteral) Equal(other Pattern) bool {
	if lit == nil {
		return other == nil
	}
	if o, ok := other.(*PatternLiteral); ok {
		return equalNodes(lit.Value, o.Value)
	}
	return false
}

func (lit *PatternLiteral) Clone() Pattern {
	if lit == nil {
		return nil
	}
	clone := *lit
	clone.Value = Clone(lit.Value)
	return &clone
}

func (lit *PatternLiteral) String() string {
	if lit == nil || lit.Value == nil {
		return "nil-literal-pattern"
	}
	return lit.Value.String()
}

type PatternVariable struct {
	PosRange
	Identifier string
}

func (*PatternVariable) patternNode() {}

func (varPat *PatternVariable) Equal(other Pattern) bool {
	if varPat == nil {
		return other == nil
	}
	if o, ok := other.(*PatternVariable); ok {
		return varPat.Identifier == o.Identifier
	}
	return false
}

func (varPat *PatternVariable) Clone() Pattern {
	if varPat == nil {
		return nil
	}
	clone := *varPat
	return &clone
}

func (varPat *PatternVariable) String() string {
	if varPat == nil {
		return "nil-variable-pattern"
	}
	return varPat.Identifier
}

type PatternWildcard struct {
	PosRange
}

func (*PatternWildcard) patternNode() {}

func (wildcard *PatternWildcard) Equal(other Pattern) bool {
	_, ok := other.(*PatternWildcard)
	return ok
}

func (wildcard *PatternWildcard) Clone() Pattern {
	if wildcard == nil {
		return nil
	}
	clone := *wildcard
	return &clone
}

func (wildcard *PatternWildcard) String() string {
	return "_"
}

type PatternSExp struct {
	PosRange
	Items []Pattern
}

func (*PatternSExp) patternNode() {}

func (sexp *PatternSExp) Equal(other Pattern) bool {
	if sexp == nil {
		return other == nil
	}
	o, ok := other.(*PatternSExp)
	if !ok {
		return false
	}
	if len(sexp.Items) != len(o.Items) {
		return false
	}
	for i := range sexp.Items {
		if !sexp.Items[i].Equal(o.Items[i]) {
			return false
		}
	}
	return true
}

func (sexp *PatternSExp) Clone() Pattern {
	if sexp == nil {
		return nil
	}
	clone := *sexp
	clone.Items = clonePatternSlice(sexp.Items)
	return &clone
}

func (sexp *PatternSExp) String() string {
	if sexp == nil {
		return "nil-sexpr-pattern"
	}
	str := &strings.Builder{}
	str.WriteRune('(')
	for i, item := range sexp.Items {
		if i > 0 {
			str.WriteRune(' ')
		}
		str.WriteString(item.String())
	}
	str.WriteRune(')')
	return str.String()
}

func clonePatternSlice(slice []Pattern) []Pattern {
	clone := make([]Pattern, len(slice))
	for i, item := range slice {
		if item != nil {
			clone[i] = item.Clone()
		}
	}
	return clone
}

type MatchCase struct {
	PosRange
	Pattern Pattern
	Body    []Node
}

func (mc *MatchCase) Clone() *MatchCase {
	if mc == nil {
		return nil
	}
	clone := *mc
	clone.Pattern = mc.Pattern.Clone()
	clone.Body = cloneSlice(mc.Body)
	return &clone
}

func (mc *MatchCase) Equal(other *MatchCase) bool {
	if mc == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	if !mc.Pattern.Equal(other.Pattern) {
		return false
	}
	return equalSlices(mc.Body, other.Body)
}

func (mc *MatchCase) String() string {
	if mc == nil {
		return "nil-match-case"
	}
	str := &strings.Builder{}
	str.WriteRune('(')
	str.WriteString(mc.Pattern.String())
	for _, item := range mc.Body {
		str.WriteRune(' ')
		str.WriteString(item.String())
	}
	str.WriteRune(')')
	return str.String()
}

type Match struct {
	PosRange
	Expr  Node
	Cases []*MatchCase
}

func (match *Match) Name() string {
	return "match"
}

func (match *Match) String() string {
	if match == nil {
		return "nil-match"
	}
	str := &strings.Builder{}
	str.WriteString("(match ")
	if match.Expr != nil {
		str.WriteString(match.Expr.String())
	}
	for _, c := range match.Cases {
		str.WriteRune(' ')
		str.WriteString(c.String())
	}
	str.WriteRune(')')
	return str.String()
}

func (match *Match) Equal(other Node) bool {
	if match == nil {
		return other == nil
	}
	o, ok := other.(*Match)
	if !ok {
		return false
	}
	if !equalNodes(match.Expr, o.Expr) {
		return false
	}
	if len(match.Cases) != len(o.Cases) {
		return false
	}
	for i := range match.Cases {
		if !match.Cases[i].Equal(o.Cases[i]) {
			return false
		}
	}
	return true
}

func (match *Match) Clone() Node {
	if match == nil {
		return nil
	}
	clone := *match
	clone.Expr = Clone(match.Expr)
	clone.Cases = cloneMatchCaseSlice(match.Cases)
	return &clone
}

func cloneMatchCaseSlice(slice []*MatchCase) []*MatchCase {
	clone := make([]*MatchCase, len(slice))
	for i, item := range slice {
		clone[i] = item.Clone()
	}
	return clone
}

func equalNodes(a, b Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(b)
}

func Clone[E Node](node E) E {
	n := Node(node)

	if n == nil {
		var empty E
		return empty
	}

	return node.Clone().(E)
}

func cloneSlice[E Node](slice []E) []E {
	clone := make([]E, len(slice))

	for i, item := range slice {
		clone[i] = Clone(item)
	}

	return clone
}

func shallow[E any](ptr *E) *E {
	if ptr == nil {
		return nil
	}

	clone := *ptr
	return &clone
}

func writeStrs(wr io.StringWriter, items ...string) {
	for _, item := range items {
		_, _ = wr.WriteString(item)
	}
}

func joinStringers[S fmt.Stringer](wr io.StringWriter, sep string, items []S) {
	if len(items) == 0 {
		return
	}

	_, _ = wr.WriteString(items[0].String())

	for _, item := range items[1:] {
		_, _ = wr.WriteString(sep)
		_, _ = wr.WriteString(item.String())
	}
}

func equalSlices[N1, N2 Node](slice []N1, other []N2) bool {
	return slices.EqualFunc(slice, other, func(n1 N1, n2 N2) bool {
		return n1.Equal(n2)
	})
}
