package cii_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

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
		DeliveryDate:   time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Seller:         einvoice.Party{Name: "Seller GmbH", CountryCode: "DE", VATID: "DE123456789", ElectronicAddress: "seller@example.de"},
		Buyer:          einvoice.Party{Name: "Buyer AG", CountryCode: "DE", ElectronicAddress: "buyer@example.de"},
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
		"<ram:URIID schemeID=\"EM\">buyer@example.de</ram:URIID>", // BT-49 buyer electronic address
		"<ram:OccurrenceDateTime>",                                // BT-72 delivery date
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

// Standard sale with a document-level discount (allowance): the discount reduces
// the taxable basis, so VAT is charged on the net-of-discount amount.
func TestSerializeAllowance(t *testing.T) {
	inv := sampleInvoice() // 10 * 100 = 1000 net @ 19%
	inv.AllowanceCharges = []einvoice.AllowanceCharge{{
		Amount: dec("100.00"), TaxCategory: einvoice.CategoryStandard, TaxRate: dec("19"), Reason: "Treuerabatt",
	}}
	out, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, want := range []string{
		"<udt:Indicator>false</udt:Indicator>", // allowance, not charge
		"<ram:ActualAmount>100.00</ram:ActualAmount>",
		"Treuerabatt",
		"<ram:AllowanceTotalAmount>100.00</ram:AllowanceTotalAmount>",
		"<ram:TaxBasisTotalAmount>900.00</ram:TaxBasisTotalAmount>",          // 1000 - 100
		"<ram:TaxTotalAmount currencyID=\"EUR\">171.00</ram:TaxTotalAmount>", // 900 * 19%
		"<ram:GrandTotalAmount>1071.00</ram:GrandTotalAmount>",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %q", want)
		}
	}
}

// New car sold tax-free to an EU business (intra-community supply, category K).
// Requires a deliver-to address (BG-15), which defaults to the buyer's.
func TestSerializeIntraCommunity(t *testing.T) {
	inv := &einvoice.Invoice{
		Number: "RE-IC-1", Currency: "EUR",
		DeliveryDate: time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
		Seller:       einvoice.Party{Name: "Autohaus GmbH", CountryCode: "DE", VATID: "DE123456789", ElectronicAddress: "s@a.de"},
		Buyer:        einvoice.Party{Name: "Garage Paris", Street: "Rue 1", City: "Paris", PostalCode: "75001", CountryCode: "FR", VATID: "FR12345678901", ElectronicAddress: "b@g.fr"},
		LineItems: []einvoice.LineItem{{
			Description: "Neuwagen", Quantity: dec("1"), UnitCode: "C62", UnitPrice: dec("25000.00"),
			TaxCategory: einvoice.CategoryIntraCommunity, TaxRate: dec("0"), ExemptionCode: einvoice.VATExIntraCommunity,
		}},
	}
	out, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, want := range []string{
		"<ram:CategoryCode>K</ram:CategoryCode>",
		"<ram:ExemptionReasonCode>VATEX-EU-IC</ram:ExemptionReasonCode>",
		"<ram:ShipToTradeParty>", // BG-15 deliver-to address
		"<ram:CountryID>FR</ram:CountryID>",
		"<ram:GrandTotalAmount>25000.00</ram:GrandTotalAmount>", // tax-free
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %q", want)
		}
	}
}

// Reverse charge (§13b): VAT owed by the recipient, category AE, rate 0.
func TestSerializeReverseCharge(t *testing.T) {
	inv := &einvoice.Invoice{
		Number: "RE-AE-1", Currency: "EUR",
		Seller: einvoice.Party{Name: "Seller", CountryCode: "DE", VATID: "DE123456789", ElectronicAddress: "s@a.de"},
		Buyer:  einvoice.Party{Name: "Buyer", CountryCode: "DE", VATID: "DE987654321", ElectronicAddress: "b@b.de"},
		LineItems: []einvoice.LineItem{{
			Description: "Leistung", Quantity: dec("1"), UnitCode: "C62", UnitPrice: dec("5000.00"),
			TaxCategory: einvoice.CategoryReverseCharge, TaxRate: dec("0"), ExemptionCode: einvoice.VATExReverseCharge,
		}},
	}
	out, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	for _, want := range []string{
		"<ram:CategoryCode>AE</ram:CategoryCode>",
		"Steuerschuldnerschaft des Leistungsempfängers", // BG-1 note + BT-120
		"<ram:GrandTotalAmount>5000.00</ram:GrandTotalAmount>",
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("output missing %q", want)
		}
	}
}
