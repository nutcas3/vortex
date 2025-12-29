package math

import "github.com/shopspring/decimal"

// Constants
var (
	Zero     = decimal.NewFromInt(0)
	One      = decimal.NewFromInt(1)
	Point5   = decimal.NewFromFloat(0.5)
	Point1   = decimal.NewFromFloat(0.1)
	Point01  = decimal.NewFromFloat(0.01)
	Point001 = decimal.NewFromFloat(0.001)
)

func Abs(x decimal.Decimal) decimal.Decimal {
	if x.LessThan(Zero) {
		return x.Neg()
	}
	return x
}

func Min(a, b decimal.Decimal) decimal.Decimal {
	if a.LessThan(b) {
		return a
	}
	return b
}

func Clamp(value, min, max decimal.Decimal) decimal.Decimal {
	if value.LessThan(min) {
		return min
	}
	if value.GreaterThan(max) {
		return max
	}
	return value
}
