package cii

import (
	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/shopspring/decimal"
)

// electronicAddress builds the party's electronic address (BT-34/BT-49). It uses
// the explicit ElectronicAddress, falling back to Contact.Email; the scheme
// defaults to "EM" (email). Returns nil when no address is available.
func electronicAddress(p einvoice.Party) *UniversalCommunication {
	addr, scheme := p.ElectronicAddress, p.ElectronicAddressScheme
	if addr == "" && p.Contact != nil {
		addr = p.Contact.Email
	}
	if addr == "" {
		return nil
	}
	if scheme == "" {
		scheme = "EM"
	}
	return &UniversalCommunication{URIID: &IDWithScheme{Value: addr, SchemeID: scheme}}
}

// fmtAmount2 formats a decimal to "123.45" (2 decimal places).
func fmtAmount2(d decimal.Decimal) string {
	return d.StringFixed(2)
}

// fmtQty formats a quantity
func fmtQty(d decimal.Decimal) string {
	return d.StringFixed(4)
}
