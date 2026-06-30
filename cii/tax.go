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

// hasCategory reports whether any line uses the given VAT category.
func hasCategory(inv *einvoice.Invoice, cat einvoice.TaxCategory) bool {
	for _, li := range inv.LineItems {
		if categoryOf(li) == string(cat) {
			return true
		}
	}
	return false
}

// taxBreakdown groups the invoice lines into VAT breakdown entries, one per
// (category, rate), summing the taxable basis and computing the tax per group.
func taxBreakdown(inv *einvoice.Invoice) []taxGroup {
	var groups []taxGroup
	index := map[string]int{}
	group := func(cat string, rate decimal.Decimal, code, reason string) int {
		key := cat + "|" + rate.String()
		i, ok := index[key]
		if !ok {
			groups = append(groups, taxGroup{category: cat, rate: rate, exCode: code, exReason: reason})
			i = len(groups) - 1
			index[key] = i
		}
		return i
	}
	for _, li := range inv.LineItems {
		code, reason := exemptionOf(li)
		i := group(categoryOf(li), li.TaxRate, code, reason)
		groups[i].basis = groups[i].basis.Add(lineNet(li))
	}
	// Document-level allowances reduce, charges add to their group's basis.
	for _, ac := range inv.AllowanceCharges {
		cat := string(ac.TaxCategory)
		if cat == "" {
			cat = string(einvoice.CategoryStandard)
		}
		i := group(cat, ac.TaxRate, "", "")
		if ac.IsCharge {
			groups[i].basis = groups[i].basis.Add(ac.Amount)
		} else {
			groups[i].basis = groups[i].basis.Sub(ac.Amount)
		}
	}
	for i := range groups {
		groups[i].tax = groups[i].basis.Mul(groups[i].rate).Div(hundred).Round(2)
	}
	return groups
}
