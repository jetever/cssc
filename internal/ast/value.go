package ast

// Value is a css value, e.g. dimension, percentage, or number.
type Value interface {
	Node

	// isValue is only used for type discrimination.
	isValue()
}

// String is a string literal.
type String struct {
	Span

	// Value is the string.
	Value string
}

// Location implements Node.
func (n *String) Location() *Span { return &n.Span }

// Dimension is a numeric value and a unit. Dimension can
// also represent Percentages (% unit) or Numbers (empty string unit).
type Dimension struct {
	Span

	// Value is the string representation for the value.
	Value string

	// Unit is the unit (e.g. rem, px) for the dimension. If Unit
	// is empty, then it's a CSS number type.
	Unit string
}

// Location implements Node.
func (n *Dimension) Location() *Span { return &n.Span }

// Percentage is a numeric percentage.
type Percentage struct {
	Span

	// Value is the string representation for the value.
	Value string
}

// Location implements Node.
func (n *Percentage) Location() *Span { return &n.Span }

// Identifier is any string identifier value, e.g. inherit or left.
type Identifier struct {
	Span

	// Value is the identifier.
	Value string
}

// Location implements Node.
func (n *Identifier) Location() *Span { return &n.Span }

// HexColor is a hex color (e.g. #aabbccdd) defined by https://www.w3.org/TR/css-color-3/.
type HexColor struct {
	Span

	// RGBA is the literal rgba value.
	RGBA string
}

// Location implements Node.
func (n *HexColor) Location() *Span { return &n.Span }

// Function is a css function.
type Function struct {
	Span

	// Name is the name of the function.
	Name string

	// Arguments is the set of values passed into the function.
	Arguments []Value
}

// Location implements Node.
func (f Function) Location() *Span { return &f.Span }

// IsMath returns whether or not this function supports math expressions
// as values.
func (f Function) IsMath() bool {
	_, ok := mathFunctions[f.Name]
	return ok
}

var mathFunctions = map[string]struct{}{
	"calc":  struct{}{},
	"min":   struct{}{},
	"max":   struct{}{},
	"clamp": struct{}{},
}

// MathExpression is a binary expression for math functions.
type MathExpression struct {
	Span

	// Operator +, -, *, or /.
	Operator string

	Left  Value
	Right Value
}

// Location implements Node.
func (n *MathExpression) Location() *Span { return &n.Span }

// Comma is a single comma. Some declarations require commas,
// e.g. font-family fallbacks or transitions.
type Comma struct {
	Span
}

// Location implements Node.
func (n *Comma) Location() *Span { return &n.Span }

func (String) isValue()         {}
func (Dimension) isValue()      {}
func (Function) isValue()       {}
func (MathExpression) isValue() {}
func (Comma) isValue()          {}
func (Identifier) isValue()     {}
func (HexColor) isValue()       {}

var _ Value = &String{}
var _ Value = &Dimension{}
var _ Value = &Function{}
var _ Value = &MathExpression{}
var _ Value = &Comma{}
var _ Value = &Identifier{}
var _ Value = &HexColor{}
