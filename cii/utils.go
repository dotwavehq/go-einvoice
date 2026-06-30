package cii

import "github.com/shopspring/decimal"

// fmtAmount2 formats a decimal to "123.45" (2 decimal places).
func fmtAmount2(d decimal.Decimal) string {
	return d.StringFixed(2)
}

// fmtQty formats a quantity
func fmtQty(d decimal.Decimal) string {
	return d.StringFixed(4)
}
