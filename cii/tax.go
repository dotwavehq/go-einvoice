package cii

import (
	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/shopspring/decimal"
)

var hundred = decimal.NewFromInt(100)

// taxGroup is one EN 16931 VAT breakdown entry (BG-23): one per unique
// combination of category code (BT-118) and rate (BT-119).
type taxGroup struct {
	category string
	rate     decimal.Decimal
	basis    decimal.Decimal
	tax      decimal.Decimal
	exCode   string
	exReason string
}

// categoryOf returns the line's VAT category, defaulting to Standard ("S").
func categoryOf(li einvoice.LineItem) string {
	if li.TaxCategory == "" {
		return string(einvoice.CategoryStandard)
	}
	return string(li.TaxCategory)
}

// exemptionOf resolves the line's exemption code (BT-121) and text (BT-120),
// falling back to the standard German text for a known code.
func exemptionOf(li einvoice.LineItem) (code, reason string) {
	code = string(li.ExemptionCode)
	reason = li.ExemptionReason
	if reason == "" && li.ExemptionCode != "" {
		reason = li.ExemptionCode.DefaultReason()
	}
	return code, reason
}

// lineNet returns the net amount of a line (BT-131): quantity * unit price.
func lineNet(li einvoice.LineItem) decimal.Decimal {
	return li.Quantity.Mul(li.UnitPrice)
}

// taxBreakdown groups the invoice lines into VAT breakdown entries, one per
// (category, rate), summing the taxable basis and computing the tax per group.
func taxBreakdown(inv *einvoice.Invoice) []taxGroup {
	var groups []taxGroup
	index := map[string]int{}
	for _, li := range inv.LineItems {
		cat := categoryOf(li)
		key := cat + "|" + li.TaxRate.String()
		i, ok := index[key]
		if !ok {
			code, reason := exemptionOf(li)
			groups = append(groups, taxGroup{category: cat, rate: li.TaxRate, exCode: code, exReason: reason})
			i = len(groups) - 1
			index[key] = i
		}
		groups[i].basis = groups[i].basis.Add(lineNet(li))
	}
	for i := range groups {
		groups[i].tax = groups[i].basis.Mul(groups[i].rate).Div(hundred).Round(2)
	}
	return groups
}
