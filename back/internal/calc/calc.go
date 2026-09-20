// Package calc implements the calculator's arithmetic operations without any
// dependency on HTTP so each operation can be unit-tested in isolation.
package calc

import (
	"errors"
	"math"
)

// Sentinel errors returned by the operations; the HTTP layer maps each one to
// a status code and an error code.
var (
	ErrDivisionByZero = errors.New("division by zero")
	ErrNegativeSqrt   = errors.New("square root of a negative number")
	ErrOverflow       = errors.New("numeric overflow")
	ErrUndefined      = errors.New("operation is undefined")
)

// Add returns a + b.
func Add(a, b float64) (float64, error) {
	return finite(a + b)
}

// Subtract returns a - b.
func Subtract(a, b float64) (float64, error) {
	return finite(a - b)
}

// Multiply returns a * b.
func Multiply(a, b float64) (float64, error) {
	return finite(a * b)
}

// Divide returns a / b, or ErrDivisionByZero when b is zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return finite(a / b)
}

// Pow returns base raised to exponent, guarding against results that are
// undefined or not representable.
func Pow(base, exponent float64) (float64, error) {
	if base == 0 && exponent < 0 {
		return 0, ErrDivisionByZero
	}
	return finite(math.Pow(base, exponent))
}

// Sqrt returns the square root of a, or ErrNegativeSqrt when a is negative.
func Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, ErrNegativeSqrt
	}
	return finite(math.Sqrt(a))
}

// Percentage returns a percent of b, that is a*b/100.
func Percentage(a, b float64) (float64, error) {
	return finite(a * b / 100)
}

// finite rejects NaN and infinite results so callers never emit numbers that
// cannot be represented as finite JSON numbers.
func finite(v float64) (float64, error) {
	switch {
	case math.IsNaN(v):
		return 0, ErrUndefined
	case math.IsInf(v, 0):
		return 0, ErrOverflow
	default:
		return v, nil
	}
}
