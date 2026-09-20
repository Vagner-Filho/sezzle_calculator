package calc_test

import (
	"errors"
	"math"
	"testing"

	"calculator/back/internal/calc"
)

type binaryCase struct {
	name    string
	a, b    float64
	want    float64
	wantErr error
}

func TestAdd(t *testing.T) {
	runBinaryTests(t, calc.Add, []binaryCase{
		{name: "positive integers", a: 2, b: 3, want: 5},
		{name: "mixed signs", a: -2, b: 3, want: 1},
		{name: "floating point", a: 0.1, b: 0.2, want: 0.30000000000000004},
		{name: "overflow", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: calc.ErrOverflow},
		{name: "not a number", a: math.Inf(1), b: math.Inf(-1), wantErr: calc.ErrUndefined},
	})
}

func TestSubtract(t *testing.T) {
	runBinaryTests(t, calc.Subtract, []binaryCase{
		{name: "positive result", a: 5, b: 3, want: 2},
		{name: "negative result", a: 3, b: 5, want: -2},
	})
}

func TestMultiply(t *testing.T) {
	runBinaryTests(t, calc.Multiply, []binaryCase{
		{name: "integers", a: 4, b: 2.5, want: 10},
		{name: "negative factor", a: -3, b: 3, want: -9},
		{name: "overflow", a: math.MaxFloat64, b: 2, wantErr: calc.ErrOverflow},
	})
}

func TestDivide(t *testing.T) {
	runBinaryTests(t, calc.Divide, []binaryCase{
		{name: "exact", a: 10, b: 4, want: 2.5},
		{name: "negative divisor", a: 7, b: -2, want: -3.5},
		{name: "division by zero", a: 1, b: 0, wantErr: calc.ErrDivisionByZero},
		{name: "zero divided by zero", a: 0, b: 0, wantErr: calc.ErrDivisionByZero},
		{name: "division by negative zero", a: 1, b: math.Copysign(0, -1), wantErr: calc.ErrDivisionByZero},
	})
}

func TestPow(t *testing.T) {
	runBinaryTests(t, calc.Pow, []binaryCase{
		{name: "positive exponent", a: 2, b: 10, want: 1024},
		{name: "fractional exponent", a: 9, b: 0.5, want: 3},
		{name: "negative exponent", a: 2, b: -1, want: 0.5},
		{name: "zero to a negative power", a: 0, b: -1, wantErr: calc.ErrDivisionByZero},
		{name: "negative base fractional exponent", a: -1, b: 0.5, wantErr: calc.ErrUndefined},
		{name: "overflow", a: 10, b: 400, wantErr: calc.ErrOverflow},
	})
}

func TestPercentage(t *testing.T) {
	runBinaryTests(t, calc.Percentage, []binaryCase{
		{name: "half of a value", a: 50, b: 200, want: 100},
		{name: "decimal result", a: 10, b: 45, want: 4.5},
		{name: "zero percent", a: 0, b: 5, want: 0},
		{name: "overflow", a: math.MaxFloat64, b: math.MaxFloat64, wantErr: calc.ErrOverflow},
	})
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr error
	}{
		{name: "perfect square", a: 9, want: 3},
		{name: "zero", a: 0, want: 0},
		{name: "irrational", a: 2, want: math.Sqrt2},
		{name: "negative input", a: -1, wantErr: calc.ErrNegativeSqrt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Sqrt(tt.a)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

// runBinaryTests exercises a two-operand operation with a shared table.
func runBinaryTests(t *testing.T, op func(a, b float64) (float64, error), tests []binaryCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := op(tt.a, tt.b)
			assertResult(t, got, err, tt.want, tt.wantErr)
		})
	}
}

func assertResult(t *testing.T, got float64, err error, want float64, wantErr error) {
	t.Helper()
	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
