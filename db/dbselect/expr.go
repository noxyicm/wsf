package dbselect

// Expr is an expression type
type Expr struct {
	expression string
}

// ToString renders an expression
func (e *Expr) ToString() string {
	return e.expression
}

// Assemble renders an expression
func (e *Expr) Assemble() string {
	return e.expression
}

// NewExpr creates a new expression
func NewExpr(expr string) *Expr {
	return &Expr{
		expression: expr,
	}
}
