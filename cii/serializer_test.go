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
		TaxTotal:   dec("190.00"),
		GrandTotal: dec("1190.00"),
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
