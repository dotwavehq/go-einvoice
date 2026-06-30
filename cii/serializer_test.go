package cii_test

import (
	"bytes"
	"strings"
	"testing"

	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/dotwavehq/go-einvoice/cii"
	"github.com/shopspring/decimal"
)

func dec(s string) decimal.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}

func sampleInvoice() *einvoice.Invoice {
	return &einvoice.Invoice{
		Number:         "RE-2025-1001",
		Currency:       "EUR",
		BuyerReference: "ORDER-12345",
		Seller:         einvoice.Party{Name: "Seller GmbH", CountryCode: "DE", VATID: "DE123456789"},
		Buyer:          einvoice.Party{Name: "Buyer AG", CountryCode: "DE"},
		Payment:        einvoice.Payment{IBAN: "DE89370400440532013000"},
		LineItems: []einvoice.LineItem{{
			Description: "Consulting",
			Quantity:    dec("10"),
			UnitCode:    "HUR",
			UnitPrice:   dec("100.00"),
			TaxRate:     dec("19"),
		}},
	}
}

func TestSerialize(t *testing.T) {
	out, err := cii.NewSerializer().Serialize(sampleInvoice())
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	if !bytes.HasPrefix(out, []byte("<?xml")) {
		t.Error("output missing XML header")
	}

	for _, want := range []string{
		"RE-2025-1001",           // invoice number
		cii.ProfileXRechnung3,    // XRechnung 3.0 profile
		"ORDER-12345",            // buyer reference
		"DE123456789",            // seller VAT id
		"DE89370400440532013000", // IBAN
		"<ram:GrandTotalAmount>1190.00</ram:GrandTotalAmount>",
		"<ram:TaxTotalAmount currencyID=\"EUR\">190.00</ram:TaxTotalAmount>",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestSerializeBuyerReferenceFallback(t *testing.T) {
	inv := sampleInvoice()
	inv.BuyerReference = ""
	out, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if !strings.Contains(string(out), "NOT_PROVIDED") {
		t.Error("empty BuyerReference should fall back to NOT_PROVIDED")
	}
}

// Car-dealer case: a margin-scheme used car (§25a, category E) plus a standard
// 19% delivery fee — a mixed-rate invoice that the single-breakdown code could
// not represent.
func TestSerializeMarginSchemeMixedRate(t *testing.T) {
	inv := &einvoice.Invoice{
		Number:   "RE-2025-2002",
		Currency: "EUR",
		Seller:   einvoice.Party{Name: "Autohaus GmbH", CountryCode: "DE", VATID: "DE123456789"},
		Buyer:    einvoice.Party{Name: "Käufer", CountryCode: "DE"},
		LineItems: []einvoice.LineItem{
			{
				Description:   "Gebrauchtwagen VW Golf",
				Quantity:      dec("1"),
				UnitCode:      "C62",
				UnitPrice:     dec("10000.00"),
				TaxCategory:   einvoice.CategoryExempt,
				TaxRate:       dec("0"),
				ExemptionCode: einvoice.VATExSecondHandGoods,
			},
			{
				Description: "Überführung",
				Quantity:    dec("1"),
				UnitCode:    "C62",
				UnitPrice:   dec("100.00"),
				TaxCategory: einvoice.CategoryStandard,
				TaxRate:     dec("19"),
			},
		},
	}

	out, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	s := string(out)

	for _, want := range []string{
		"<ram:CategoryCode>E</ram:CategoryCode>", // margin-scheme group
		"<ram:CategoryCode>S</ram:CategoryCode>", // standard 19% group
		"<ram:ExemptionReasonCode>VATEX-EU-F</ram:ExemptionReasonCode>",
		"Gebrauchtgegenstände/Sonderregelung",                   // BT-120 + BG-1 note
		"<ram:GrandTotalAmount>10119.00</ram:GrandTotalAmount>", // 10000 + 100 + 19 VAT
		"<ram:TaxTotalAmount currencyID=\"EUR\">19.00</ram:TaxTotalAmount>",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q", want)
		}
	}

	// VAT must not be charged on the margin-scheme line: its breakdown tax is 0.
	if !strings.Contains(s, "<ram:CalculatedAmount>0.00</ram:CalculatedAmount>") {
		t.Error("margin-scheme group must have CalculatedAmount 0.00")
	}
}
