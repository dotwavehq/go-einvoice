package einvoice

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ParsedInvoice is the result of reading an e-invoice XML. It carries the
// canonical Invoice plus the document-level monetary totals and the VAT
// breakdown as stated in the document — parsing never recomputes amounts,
// so a consumer can compare stated vs. derived values for validation.
type ParsedInvoice struct {
	Invoice

	Totals Totals
	// Taxes is the document-level VAT breakdown (BG-23).
	Taxes []TaxBreakdown
	// Format is the detected syntax: "cii" (ZUGFeRD / Factur-X / XRechnung
	// CII) or "ubl" (XRechnung UBL, Peppol BIS).
	Format string
	// Guideline is the specification identifier (BT-24): the CII guideline
	// URN or the UBL CustomizationID.
	Guideline string
}

// Totals are the document-level amounts (BG-22) as stated in the XML.
type Totals struct {
	LineTotal     decimal.Decimal // BT-106
	TaxBasisTotal decimal.Decimal // BT-109
	TaxTotal      decimal.Decimal // BT-110
	GrandTotal    decimal.Decimal // BT-112
	DuePayable    decimal.Decimal // BT-115
}

// TaxBreakdown is one VAT category subtotal (BG-23).
type TaxBreakdown struct {
	CategoryCode string
	Rate         decimal.Decimal
	Basis        decimal.Decimal
	Amount       decimal.Decimal
}

// Parse reads an EN 16931 e-invoice XML in either CII or UBL syntax.
// It is lenient: fields absent from the document stay zero — callers that
// need guarantees should check the fields they rely on.
func Parse(data []byte) (*ParsedInvoice, error) {
	root, err := rootLocalName(data)
	if err != nil {
		return nil, fmt.Errorf("einvoice: not an XML document: %w", err)
	}
	switch root {
	case "CrossIndustryInvoice":
		return parseCII(data)
	case "Invoice", "CreditNote":
		return parseUBL(data, root)
	default:
		return nil, fmt.Errorf("einvoice: unsupported root element %q", root)
	}
}

func rootLocalName(data []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", err
		}
		if start, ok := tok.(xml.StartElement); ok {
			return start.Name.Local, nil
		}
	}
}

// Decode structs use unqualified names on purpose: encoding/xml then
// matches by local name regardless of namespace prefix, which the
// serializer-side structs (literal "ram:" prefixes) cannot do.

type ciiDoc struct {
	Guideline      string    `xml:"ExchangedDocumentContext>GuidelineSpecifiedDocumentContextParameter>ID"`
	Number         string    `xml:"ExchangedDocument>ID"`
	TypeCode       string    `xml:"ExchangedDocument>TypeCode"`
	IssueDate      ciiDate   `xml:"ExchangedDocument>IssueDateTime>DateTimeString"`
	Notes          []string  `xml:"ExchangedDocument>IncludedNote>Content"`
	BuyerReference string    `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeAgreement>BuyerReference"`
	Seller         ciiParty  `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeAgreement>SellerTradeParty"`
	Buyer          ciiParty  `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeAgreement>BuyerTradeParty"`
	DeliveryDate   ciiDate   `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeDelivery>ActualDeliverySupplyChainEvent>OccurrenceDateTime>DateTimeString"`
	Lines          []ciiLine `xml:"SupplyChainTradeTransaction>IncludedSupplyChainTradeLineItem"`

	Currency         string               `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>InvoiceCurrencyCode"`
	PaymentMeans     []ciiPaymentMeans    `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>SpecifiedTradeSettlementPaymentMeans"`
	TradeTaxes       []ciiTradeTax        `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>ApplicableTradeTax"`
	AllowanceCharges []ciiAllowanceCharge `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>SpecifiedTradeAllowanceCharge"`
	DueDate          ciiDate              `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>SpecifiedTradePaymentTerms>DueDateDateTime>DateTimeString"`
	Summation        ciiSummation         `xml:"SupplyChainTradeTransaction>ApplicableHeaderTradeSettlement>SpecifiedTradeSettlementHeaderMonetarySummation"`
}

type ciiAllowanceCharge struct {
	IsCharge     bool   `xml:"ChargeIndicator>Indicator"`
	Amount       string `xml:"ActualAmount"`
	Reason       string `xml:"Reason"`
	CategoryCode string `xml:"CategoryTradeTax>CategoryCode"`
	Rate         string `xml:"CategoryTradeTax>RateApplicablePercent"`
}

type ciiDate struct {
	Value  string `xml:",chardata"`
	Format string `xml:"format,attr"`
}

type ciiParty struct {
	Name          string `xml:"Name"`
	ContactPerson string `xml:"DefinedTradeContact>PersonName"`
	ContactPhone  string `xml:"DefinedTradeContact>TelephoneUniversalCommunication>CompleteNumber"`
	ContactEmail  string `xml:"DefinedTradeContact>EmailURIUniversalCommunication>URIID"`
	Street        string `xml:"PostalTradeAddress>LineOne"`
	Postcode      string `xml:"PostalTradeAddress>PostcodeCode"`
	City          string `xml:"PostalTradeAddress>CityName"`
	Country       string `xml:"PostalTradeAddress>CountryID"`
	ElectronicURI struct {
		Value    string `xml:",chardata"`
		SchemeID string `xml:"schemeID,attr"`
	} `xml:"URIUniversalCommunication>URIID"`
	TaxRegistrations []struct {
		Value    string `xml:",chardata"`
		SchemeID string `xml:"schemeID,attr"`
	} `xml:"SpecifiedTaxRegistration>ID"`
}

type ciiPaymentMeans struct {
	TypeCode      string `xml:"TypeCode"`
	IBAN          string `xml:"PayeePartyCreditorFinancialAccount>IBANID"`
	AccountHolder string `xml:"PayeePartyCreditorFinancialAccount>AccountName"`
	BIC           string `xml:"PayeeSpecifiedCreditorFinancialInstitution>BICID"`
}

type ciiTradeTax struct {
	Amount       string `xml:"CalculatedAmount"`
	Basis        string `xml:"BasisAmount"`
	CategoryCode string `xml:"CategoryCode"`
	Rate         string `xml:"RateApplicablePercent"`
}

type ciiSummation struct {
	LineTotal     string `xml:"LineTotalAmount"`
	TaxBasisTotal string `xml:"TaxBasisTotalAmount"`
	TaxTotal      string `xml:"TaxTotalAmount"`
	GrandTotal    string `xml:"GrandTotalAmount"`
	DuePayable    string `xml:"DuePayableAmount"`
}

type ciiLine struct {
	Name            string `xml:"SpecifiedTradeProduct>Name"`
	UnitPrice       string `xml:"SpecifiedLineTradeAgreement>NetPriceProductTradePrice>ChargeAmount"`
	Quantity        ciiQty `xml:"SpecifiedLineTradeDelivery>BilledQuantity"`
	TaxCategory     string `xml:"SpecifiedLineTradeSettlement>ApplicableTradeTax>CategoryCode"`
	TaxRate         string `xml:"SpecifiedLineTradeSettlement>ApplicableTradeTax>RateApplicablePercent"`
	ExemptionCode   string `xml:"SpecifiedLineTradeSettlement>ApplicableTradeTax>ExemptionReasonCode"`
	ExemptionReason string `xml:"SpecifiedLineTradeSettlement>ApplicableTradeTax>ExemptionReason"`
}

type ciiQty struct {
	Value    string `xml:",chardata"`
	UnitCode string `xml:"unitCode,attr"`
}

func parseCII(data []byte) (*ParsedInvoice, error) {
	var doc ciiDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("einvoice: parse cii: %w", err)
	}

	p := &ParsedInvoice{Format: "cii", Guideline: strings.TrimSpace(doc.Guideline)}
	p.Number = doc.Number
	p.TypeCode = doc.TypeCode
	p.IssueDate = parseCIIDate(doc.IssueDate)
	p.DueDate = parseCIIDate(doc.DueDate)
	p.DeliveryDate = parseCIIDate(doc.DeliveryDate)
	p.Currency = doc.Currency
	p.Note = strings.Join(doc.Notes, "\n")
	p.BuyerReference = doc.BuyerReference
	p.Seller = ciiPartyToParty(doc.Seller)
	p.Buyer = ciiPartyToParty(doc.Buyer)

	for _, pm := range doc.PaymentMeans {
		applyPaymentMeans(&p.Payment, pm.TypeCode, pm.IBAN, pm.BIC, pm.AccountHolder)
	}

	for _, tt := range doc.TradeTaxes {
		p.Taxes = append(p.Taxes, TaxBreakdown{
			CategoryCode: tt.CategoryCode,
			Rate:         dec(tt.Rate),
			Basis:        dec(tt.Basis),
			Amount:       dec(tt.Amount),
		})
	}
	for _, ac := range doc.AllowanceCharges {
		p.AllowanceCharges = append(p.AllowanceCharges,
			allowanceCharge(ac.Amount, ac.IsCharge, ac.Reason, ac.CategoryCode, ac.Rate))
	}
	p.Totals = Totals{
		LineTotal:     dec(doc.Summation.LineTotal),
		TaxBasisTotal: dec(doc.Summation.TaxBasisTotal),
		TaxTotal:      dec(doc.Summation.TaxTotal),
		GrandTotal:    dec(doc.Summation.GrandTotal),
		DuePayable:    dec(doc.Summation.DuePayable),
	}

	for _, l := range doc.Lines {
		p.LineItems = append(p.LineItems, LineItem{
			Description:     l.Name,
			Quantity:        dec(l.Quantity.Value),
			UnitCode:        l.Quantity.UnitCode,
			UnitPrice:       dec(l.UnitPrice),
			TaxCategory:     TaxCategory(l.TaxCategory),
			TaxRate:         dec(l.TaxRate),
			ExemptionCode:   VATExCode(l.ExemptionCode),
			ExemptionReason: l.ExemptionReason,
		})
	}
	return p, nil
}

func ciiPartyToParty(cp ciiParty) Party {
	party := Party{
		Name:                    cp.Name,
		Street:                  cp.Street,
		City:                    cp.City,
		PostalCode:              cp.Postcode,
		CountryCode:             cp.Country,
		ElectronicAddress:       cp.ElectronicURI.Value,
		ElectronicAddressScheme: cp.ElectronicURI.SchemeID,
	}
	for _, reg := range cp.TaxRegistrations {
		// "VA" is the VAT id (BT-31); "FC" (national tax number) is ignored.
		if reg.SchemeID == "VA" {
			party.VATID = reg.Value
		}
	}
	if cp.ContactPerson != "" || cp.ContactPhone != "" || cp.ContactEmail != "" {
		party.Contact = &Contact{Name: cp.ContactPerson, Phone: cp.ContactPhone, Email: cp.ContactEmail}
	}
	return party
}

// parseCIIDate handles the UNTDID 2379 formats used by EN 16931 ("102" =
// CCYYMMDD) plus ISO dates as a fallback for lenient real-world input.
func parseCIIDate(d ciiDate) time.Time {
	v := strings.TrimSpace(d.Value)
	if v == "" {
		return time.Time{}
	}
	for _, layout := range []string{"20060102", "2006-01-02"} {
		if t, err := time.Parse(layout, v); err == nil {
			return t
		}
	}
	return time.Time{}
}

type ublDoc struct {
	CustomizationID string   `xml:"CustomizationID"`
	Number          string   `xml:"ID"`
	IssueDate       string   `xml:"IssueDate"`
	DueDate         string   `xml:"DueDate"`
	InvoiceType     string   `xml:"InvoiceTypeCode"`
	CreditNoteType  string   `xml:"CreditNoteTypeCode"`
	Notes           []string `xml:"Note"`
	Currency        string   `xml:"DocumentCurrencyCode"`
	BuyerReference  string   `xml:"BuyerReference"`
	DeliveryDate    string   `xml:"Delivery>ActualDeliveryDate"`

	Supplier         ublParty             `xml:"AccountingSupplierParty>Party"`
	Customer         ublParty             `xml:"AccountingCustomerParty>Party"`
	PaymentMeans     []ublPaymentMeans    `xml:"PaymentMeans"`
	AllowanceCharges []ublAllowanceCharge `xml:"AllowanceCharge"`
	TaxTotals        []ublTaxTotal        `xml:"TaxTotal"`
	Total            ublTotal             `xml:"LegalMonetaryTotal"`
	InvoiceLines     []ublLine            `xml:"InvoiceLine"`
	CreditLines      []ublLine            `xml:"CreditNoteLine"`
}

type ublParty struct {
	EndpointID struct {
		Value    string `xml:",chardata"`
		SchemeID string `xml:"schemeID,attr"`
	} `xml:"EndpointID"`
	Name             string `xml:"PartyName>Name"`
	RegistrationName string `xml:"PartyLegalEntity>RegistrationName"`
	Street           string `xml:"PostalAddress>StreetName"`
	City             string `xml:"PostalAddress>CityName"`
	Postcode         string `xml:"PostalAddress>PostalZone"`
	Country          string `xml:"PostalAddress>Country>IdentificationCode"`
	VATID            string `xml:"PartyTaxScheme>CompanyID"`
	ContactName      string `xml:"Contact>Name"`
	ContactPhone     string `xml:"Contact>Telephone"`
	ContactEmail     string `xml:"Contact>ElectronicMail"`
}

type ublPaymentMeans struct {
	Code string `xml:"PaymentMeansCode"`
	// DueDate is the CreditNote location of the due date (UBL CreditNote
	// has no top-level DueDate element).
	DueDate       string `xml:"PaymentDueDate"`
	IBAN          string `xml:"PayeeFinancialAccount>ID"`
	AccountHolder string `xml:"PayeeFinancialAccount>Name"`
	BIC           string `xml:"PayeeFinancialAccount>FinancialInstitutionBranch>ID"`
}

type ublAllowanceCharge struct {
	IsCharge     bool   `xml:"ChargeIndicator"`
	Reason       string `xml:"AllowanceChargeReason"`
	Amount       string `xml:"Amount"`
	CategoryCode string `xml:"TaxCategory>ID"`
	Rate         string `xml:"TaxCategory>Percent"`
}

type ublTaxTotal struct {
	Amount    string `xml:"TaxAmount"`
	Subtotals []struct {
		Basis        string `xml:"TaxableAmount"`
		Amount       string `xml:"TaxAmount"`
		CategoryCode string `xml:"TaxCategory>ID"`
		Rate         string `xml:"TaxCategory>Percent"`
	} `xml:"TaxSubtotal"`
}

type ublTotal struct {
	LineTotal     string `xml:"LineExtensionAmount"`
	TaxBasisTotal string `xml:"TaxExclusiveAmount"`
	GrandTotal    string `xml:"TaxInclusiveAmount"`
	DuePayable    string `xml:"PayableAmount"`
}

type ublLine struct {
	Name             string `xml:"Item>Name"`
	InvoicedQuantity ciiQty `xml:"InvoicedQuantity"`
	CreditedQuantity ciiQty `xml:"CreditedQuantity"`
	UnitPrice        string `xml:"Price>PriceAmount"`
	TaxCategory      string `xml:"Item>ClassifiedTaxCategory>ID"`
	TaxRate          string `xml:"Item>ClassifiedTaxCategory>Percent"`
}

func parseUBL(data []byte, root string) (*ParsedInvoice, error) {
	var doc ublDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("einvoice: parse ubl: %w", err)
	}

	p := &ParsedInvoice{Format: "ubl", Guideline: strings.TrimSpace(doc.CustomizationID)}
	p.Number = doc.Number
	p.TypeCode = doc.InvoiceType
	if root == "CreditNote" {
		p.TypeCode = doc.CreditNoteType
		if p.TypeCode == "" {
			p.TypeCode = TypeCreditNote
		}
	}
	if p.TypeCode == "" {
		p.TypeCode = TypeCommercialInvoice
	}
	p.IssueDate = parseISODate(doc.IssueDate)
	p.DueDate = parseISODate(doc.DueDate)
	p.DeliveryDate = parseISODate(doc.DeliveryDate)
	p.Currency = doc.Currency
	p.Note = strings.Join(doc.Notes, "\n")
	p.BuyerReference = doc.BuyerReference
	p.Seller = ublPartyToParty(doc.Supplier)
	p.Buyer = ublPartyToParty(doc.Customer)

	for _, pm := range doc.PaymentMeans {
		if p.DueDate.IsZero() {
			p.DueDate = parseISODate(pm.DueDate)
		}
		applyPaymentMeans(&p.Payment, pm.Code, pm.IBAN, pm.BIC, pm.AccountHolder)
	}

	for _, ac := range doc.AllowanceCharges {
		p.AllowanceCharges = append(p.AllowanceCharges,
			allowanceCharge(ac.Amount, ac.IsCharge, ac.Reason, ac.CategoryCode, ac.Rate))
	}

	var taxTotal decimal.Decimal
	for _, tt := range doc.TaxTotals {
		// A second TaxTotal only restates the amount in the tax currency
		// (BT-111) and never carries subtotals — summing is safe.
		for _, sub := range tt.Subtotals {
			p.Taxes = append(p.Taxes, TaxBreakdown{
				CategoryCode: sub.CategoryCode,
				Rate:         dec(sub.Rate),
				Basis:        dec(sub.Basis),
				Amount:       dec(sub.Amount),
			})
		}
		if taxTotal.IsZero() {
			taxTotal = dec(tt.Amount)
		}
	}
	p.Totals = Totals{
		LineTotal:     dec(doc.Total.LineTotal),
		TaxBasisTotal: dec(doc.Total.TaxBasisTotal),
		TaxTotal:      taxTotal,
		GrandTotal:    dec(doc.Total.GrandTotal),
		DuePayable:    dec(doc.Total.DuePayable),
	}

	lines := doc.InvoiceLines
	if root == "CreditNote" {
		lines = doc.CreditLines
	}
	for _, l := range lines {
		qty := l.InvoicedQuantity
		if qty.Value == "" {
			qty = l.CreditedQuantity
		}
		p.LineItems = append(p.LineItems, LineItem{
			Description: l.Name,
			Quantity:    dec(qty.Value),
			UnitCode:    qty.UnitCode,
			UnitPrice:   dec(l.UnitPrice),
			TaxCategory: TaxCategory(l.TaxCategory),
			TaxRate:     dec(l.TaxRate),
		})
	}
	return p, nil
}

func ublPartyToParty(up ublParty) Party {
	name := up.Name
	if name == "" {
		name = up.RegistrationName
	}
	party := Party{
		Name:                    name,
		Street:                  up.Street,
		City:                    up.City,
		PostalCode:              up.Postcode,
		CountryCode:             up.Country,
		VATID:                   up.VATID,
		ElectronicAddress:       up.EndpointID.Value,
		ElectronicAddressScheme: up.EndpointID.SchemeID,
	}
	if up.ContactName != "" || up.ContactPhone != "" || up.ContactEmail != "" {
		party.Contact = &Contact{Name: up.ContactName, Phone: up.ContactPhone, Email: up.ContactEmail}
	}
	return party
}

func parseISODate(v string) time.Time {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Time{}
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}
	}
	return t
}

// applyPaymentMeans folds one payment-means entry into Payment: the first
// entry with an IBAN becomes the primary account, further ones land in
// AdditionalAccounts (BG-17 is 0..n).
func applyPaymentMeans(p *Payment, code, iban, bic, holder string) {
	if p.PaymentMeansCode == "" {
		p.PaymentMeansCode = code
	}
	if iban == "" {
		return
	}
	if p.IBAN == "" {
		p.IBAN, p.BIC, p.AccountHolder = iban, bic, holder
		return
	}
	p.AdditionalAccounts = append(p.AdditionalAccounts,
		PaymentAccount{IBAN: iban, BIC: bic, AccountHolder: holder})
}

func allowanceCharge(amount string, isCharge bool, reason, category, rate string) AllowanceCharge {
	return AllowanceCharge{
		Amount:      dec(amount),
		IsCharge:    isCharge,
		TaxCategory: TaxCategory(category),
		TaxRate:     dec(rate),
		Reason:      reason,
	}
}

// dec parses a decimal, mapping empty/malformed input to zero — parse
// leniency over hard failure, stated amounts are validated by consumers.
func dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(strings.TrimSpace(v))
	if err != nil {
		return decimal.Decimal{}
	}
	return d
}
