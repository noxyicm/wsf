package db

// SQLExpr represents an SQL expression
type SQLExpr struct {
	expression string
}

// ToString renders an expression
func (e *SQLExpr) ToString() string {
	return e.expression
}

// Assemble renders an expression
func (e *SQLExpr) Assemble() string {
	return e.expression
}

// NewExpr creates a new expression
func NewExpr(expr string) *SQLExpr {
	return &SQLExpr{
		expression: expr,
	}
}
