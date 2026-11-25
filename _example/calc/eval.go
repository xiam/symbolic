package calc

import "fmt"

// Eval returns the value of the tree under n. It follows the rules of
// float64: a result too large is an infinity. A division by zero is an
// *Error with ErrDivisionByZero at the position of the operator.
func Eval(n Node) (float64, error) {
	switch n := n.(type) {
	case *Number:
		return n.Value, nil

	case *Paren:
		return Eval(n.X)

	case *Unary:
		x, err := Eval(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case Add:
			return x, nil
		case Sub:
			return -x, nil
		}
		return 0, fmt.Errorf("calc: invalid unary operator %q", n.Op)

	case *Binary:
		x, err := Eval(n.X)
		if err != nil {
			return 0, err
		}
		y, err := Eval(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case Add:
			return x + y, nil
		case Sub:
			return x - y, nil
		case Mul:
			return x * y, nil
		case Div:
			if y == 0 {
				return 0, &Error{Err: ErrDivisionByZero, Pos: n.OpPos}
			}
			return x / y, nil
		}
		return 0, fmt.Errorf("calc: invalid binary operator %q", n.Op)
	}

	return 0, fmt.Errorf("calc: cannot evaluate %T", n)
}
