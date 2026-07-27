package einvoice_test

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"

	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/dotwavehq/go-einvoice/cii"
)

// TestParseCIIRoundTrip serializes an invoice with the CII serializer and
// parses it back — parser and serializer validate each other.
func TestParseCIIRoundTrip(t *testing.T) {
	inv := &einvoice.Invoice{
		Number:         "RE-2026-0815",
		IssueDate:      time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC),
		DueDate:        time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC),
		DeliveryDate:   time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
		Currency:       "EUR",
		BuyerReference: "04011000-1234-56",
		Seller: einvoice.Party{
			Name: "Autohaus Beispiel GmbH", Street: "Musterstr. 1", City: "Köln",
			PostalCode: "50667", CountryCode: "DE", VATID: "DE123456789",
			Contact: &einvoice.Contact{Name: "E. Muster", Phone: "+49 221 123", Email: "info@beispiel.de"},
		},
		Buyer: einvoice.Party{
			Name: "Käufer AG", Street: "Weg 2", City: "Bonn",
			PostalCode: "53111", CountryCode: "DE",
		},
		Payment: einvoice.Payment{
			IBAN: "DE02120300000000202051", BIC: "BYLADEM1001",
			AccountHolder: "Autohaus Beispiel GmbH", PaymentMeansCode: "58",
		},
		LineItems: []einvoice.LineItem{
			{
				Description: "Inspektion", Quantity: decimal.NewFromInt(1), UnitCode: "C62",
				UnitPrice: decimal.NewFromFloat(200), TaxRate: decimal.NewFromInt(19),
			},
		},
		AllowanceCharges: []einvoice.AllowanceCharge{
			{Amount: decimal.NewFromInt(20), IsCharge: false, Reason: "Treuerabatt",
				TaxCategory: einvoice.CategoryStandard, TaxRate: decimal.NewFromInt(19)},
		},
	}
	xml, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		t.Fatalf("serialize: %v", err)
	}

	p, err := einvoice.Parse(xml)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.Format != "cii" {
		t.Errorf("format = %q", p.Format)
	}
	if p.Number != inv.Number || p.TypeCode != einvoice.TypeCommercialInvoice {
		t.Errorf("number/type = %q/%q", p.Number, p.TypeCode)
	}
	if !p.IssueDate.Equal(inv.IssueDate) || !p.DueDate.Equal(inv.DueDate) || !p.DeliveryDate.Equal(inv.DeliveryDate) {
		t.Errorf("dates = %v / %v / %v", p.IssueDate, p.DueDate, p.DeliveryDate)
	}
	if p.Seller.Name != inv.Seller.Name || p.Seller.VATID != inv.Seller.VATID || p.Seller.City != "Köln" {
		t.Errorf("seller = %+v", p.Seller)
	}
	if p.Payment.IBAN != inv.Payment.IBAN {
		t.Errorf("iban = %q", p.Payment.IBAN)
	}
	if p.Currency != "EUR" || p.BuyerReference != inv.BuyerReference {
		t.Errorf("currency/buyerref = %q/%q", p.Currency, p.BuyerReference)
	}
	// 200 net − 20 allowance + 19 % = 214.20 gross, stated by the
	// serializer's summation.
	if !p.Totals.GrandTotal.Equal(decimal.NewFromFloat(214.20)) || !p.Totals.TaxTotal.Equal(decimal.NewFromFloat(34.20)) {
		t.Errorf("totals = %+v", p.Totals)
	}
	if len(p.Taxes) != 1 || !p.Taxes[0].Rate.Equal(decimal.NewFromInt(19)) {
		t.Errorf("taxes = %+v", p.Taxes)
	}
	if len(p.AllowanceCharges) != 1 || p.AllowanceCharges[0].IsCharge ||
		!p.AllowanceCharges[0].Amount.Equal(decimal.NewFromInt(20)) ||
		p.AllowanceCharges[0].Reason != "Treuerabatt" {
		t.Errorf("allowance charges = %+v", p.AllowanceCharges)
	}
	if len(p.LineItems) != 1 || p.LineItems[0].Description != "Inspektion" || !p.LineItems[0].UnitPrice.Equal(decimal.NewFromInt(200)) {
		t.Errorf("lines = %+v", p.LineItems)
	}
}

const ublFixture = `<?xml version="1.0" encoding="UTF-8"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
  xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
  xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2">
  <cbc:CustomizationID>urn:cen.eu:en16931:2017#compliant#urn:xeinkauf.de:kosit:xrechnung_3.0</cbc:CustomizationID>
  <cbc:ID>4711-2026</cbc:ID>
  <cbc:IssueDate>2026-07-21</cbc:IssueDate>
  <cbc:DueDate>2026-08-04</cbc:DueDate>
  <cbc:InvoiceTypeCode>380</cbc:InvoiceTypeCode>
  <cbc:DocumentCurrencyCode>EUR</cbc:DocumentCurrencyCode>
  <cbc:BuyerReference>991-01234-56</cbc:BuyerReference>
  <cac:AccountingSupplierParty><cac:Party>
    <cbc:EndpointID schemeID="EM">rechnung@vodafone.de</cbc:EndpointID>
    <cac:PostalAddress><cbc:StreetName>Platz 1</cbc:StreetName><cbc:CityName>Düsseldorf</cbc:CityName>
      <cbc:PostalZone>40549</cbc:PostalZone><cac:Country><cbc:IdentificationCode>DE</cbc:IdentificationCode></cac:Country>
    </cac:PostalAddress>
    <cac:PartyTaxScheme><cbc:CompanyID>222/5678/9012</cbc:CompanyID><cac:TaxScheme><cbc:ID>FC</cbc:ID></cac:TaxScheme></cac:PartyTaxScheme>
    <cac:PartyTaxScheme><cbc:CompanyID>DE811223344</cbc:CompanyID><cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme></cac:PartyTaxScheme>
    <cac:PartyLegalEntity><cbc:RegistrationName>Vodafone GmbH</cbc:RegistrationName></cac:PartyLegalEntity>
  </cac:Party></cac:AccountingSupplierParty>
  <cac:AccountingCustomerParty><cac:Party>
    <cac:PartyLegalEntity><cbc:RegistrationName>Autohaus Beispiel GmbH</cbc:RegistrationName></cac:PartyLegalEntity>
  </cac:Party></cac:AccountingCustomerParty>
  <cac:PaymentMeans>
    <cbc:PaymentMeansCode>58</cbc:PaymentMeansCode>
    <cac:PayeeFinancialAccount><cbc:ID>DE75512108001245126199</cbc:ID><cbc:Name>Vodafone GmbH</cbc:Name>
      <cac:FinancialInstitutionBranch><cbc:ID>SOGEDEFFXXX</cbc:ID></cac:FinancialInstitutionBranch>
    </cac:PayeeFinancialAccount>
  </cac:PaymentMeans>
  <cac:TaxTotal><cbc:TaxAmount currencyID="EUR">15.96</cbc:TaxAmount>
    <cac:TaxSubtotal><cbc:TaxableAmount currencyID="EUR">84.00</cbc:TaxableAmount>
      <cbc:TaxAmount currencyID="EUR">15.96</cbc:TaxAmount>
      <cac:TaxCategory><cbc:ID>S</cbc:ID><cbc:Percent>19</cbc:Percent><cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme></cac:TaxCategory>
    </cac:TaxSubtotal>
  </cac:TaxTotal>
  <cac:LegalMonetaryTotal>
    <cbc:LineExtensionAmount currencyID="EUR">84.00</cbc:LineExtensionAmount>
    <cbc:TaxExclusiveAmount currencyID="EUR">84.00</cbc:TaxExclusiveAmount>
    <cbc:TaxInclusiveAmount currencyID="EUR">99.96</cbc:TaxInclusiveAmount>
    <cbc:PayableAmount currencyID="EUR">99.96</cbc:PayableAmount>
  </cac:LegalMonetaryTotal>
  <cac:InvoiceLine>
    <cbc:ID>1</cbc:ID>
    <cbc:InvoicedQuantity unitCode="C62">1</cbc:InvoicedQuantity>
    <cbc:LineExtensionAmount currencyID="EUR">84.00</cbc:LineExtensionAmount>
    <cac:Item><cbc:Name>Mobilfunk Business Juli</cbc:Name>
      <cac:ClassifiedTaxCategory><cbc:ID>S</cbc:ID><cbc:Percent>19</cbc:Percent><cac:TaxScheme><cbc:ID>VAT</cbc:ID></cac:TaxScheme></cac:ClassifiedTaxCategory>
    </cac:Item>
    <cac:Price><cbc:PriceAmount currencyID="EUR">84.00</cbc:PriceAmount></cac:Price>
  </cac:InvoiceLine>
</Invoice>`

func TestParseUBL(t *testing.T) {
	p, err := einvoice.Parse([]byte(ublFixture))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if p.Format != "ubl" || p.TypeCode != "380" {
		t.Errorf("format/type = %q/%q", p.Format, p.TypeCode)
	}
	if p.Number != "4711-2026" || p.Currency != "EUR" {
		t.Errorf("number/currency = %q/%q", p.Number, p.Currency)
	}
	if p.Seller.Name != "Vodafone GmbH" || p.Seller.VATID != "DE811223344" || p.Seller.City != "Düsseldorf" {
		t.Errorf("seller = %+v", p.Seller)
	}
	if p.Payment.IBAN != "DE75512108001245126199" || p.Payment.BIC != "SOGEDEFFXXX" {
		t.Errorf("payment = %+v", p.Payment)
	}
	if !p.DueDate.Equal(time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("due date = %v", p.DueDate)
	}
	if !p.Totals.GrandTotal.Equal(decimal.NewFromFloat(99.96)) || !p.Totals.TaxTotal.Equal(decimal.NewFromFloat(15.96)) {
		t.Errorf("totals = %+v", p.Totals)
	}
	if len(p.Taxes) != 1 || !p.Taxes[0].Rate.Equal(decimal.NewFromInt(19)) || p.Taxes[0].CategoryCode != "S" {
		t.Errorf("taxes = %+v", p.Taxes)
	}
	if len(p.LineItems) != 1 || p.LineItems[0].Description != "Mobilfunk Business Juli" {
		t.Errorf("lines = %+v", p.LineItems)
	}
}

func TestParseRejectsNonInvoiceInput(t *testing.T) {
	if _, err := einvoice.Parse([]byte("%PDF-1.7 not xml")); err == nil {
		t.Error("binary input must error")
	}
	if _, err := einvoice.Parse([]byte("<html></html>")); err == nil {
		t.Error("unknown root must error")
	}
}
