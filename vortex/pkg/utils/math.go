package utils

import "github.com/shopspring/decimal"

func Abs(x decimal.Decimal) decimal.Decimal {
	if x.LessThan(decimal.Zero) {
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
