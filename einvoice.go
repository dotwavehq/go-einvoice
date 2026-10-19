package einvoice

import (
	"time"

	"github.com/shopspring/decimal"
)

// Document type codes (BT-3, UNTDID 1001 subset permitted by EN 16931).
const (
	// TypeCommercialInvoice is a regular invoice (default).
	TypeCommercialInvoice = "380"
	// TypeCreditNote is a credit note (Gutschrift). Note the CII sign
	// convention: amounts on a 381 document are positive.
	TypeCreditNote = "381"
	// TypeCorrectedInvoice is a corrected invoice (Rechnungskorrektur /
	// Stornorechnung) carrying negative amounts that reverse the original.
	TypeCorrectedInvoice = "384"
)

type Invoice struct {
	Number string
	// TypeCode is the document type (BT-3). Empty defaults to
	// TypeCommercialInvoice ("380").
	TypeCode     string
	IssueDate    time.Time
	DueDate      time.Time
	DeliveryDate time.Time // BT-72 Leistungs-/Lieferdatum (UStG mandatory)
	// DeliveryCountryCode is the deliver-to country (BT-80). Mandatory for
	// intra-community supply (category K); defaults to the buyer's country.
	DeliveryCountryCode string
	Currency            string
	Note                string
	BuyerReference      string

	Seller           Party
	Buyer            Party
	Payment          Payment
	LineItems        []LineItem
	AllowanceCharges []AllowanceCharge // document-level allowances/charges (BG-20/21)
}

type Party struct {
	Name        string
	Street      string
	City        string
	PostalCode  string
	CountryCode string
	VATID       string
	Contact     *Contact

	// ElectronicAddress is the party's electronic address (BT-34 seller / BT-49
	// buyer), mandatory in XRechnung. ElectronicAddressScheme is its EAS code
	// (e.g. "EM" email, "9930" German VAT, "0204" Leitweg-ID); defaults to "EM".
	// If ElectronicAddress is empty, Contact.Email is used as an "EM" fallback.
	ElectronicAddress       string
	ElectronicAddressScheme string
}

type Contact struct {
	Name  string
	Phone string
	Email string
}

type Payment struct {
	IBAN             string
	BIC              string
	AccountHolder    string
	PaymentMeansCode string

	// AdditionalAccounts are further payee accounts. EN 16931 allows several
	// credit-transfer accounts (BG-17 is 0..n); each account is serialized as
	// its own SpecifiedTradeSettlementPaymentMeans with the same type code.
	AdditionalAccounts []PaymentAccount
}

// PaymentAccount is one additional payee bank account (BG-17 CREDIT TRANSFER).
type PaymentAccount struct {
	IBAN          string
	BIC           string
	AccountHolder string
}

type LineItem struct {
	Description string
	Quantity    decimal.Decimal
	UnitCode    string
	UnitPrice   decimal.Decimal

	// TaxCategory is the EN 16931 VAT category (BT-151). Empty defaults to
	// Standard ("S"). For Differenzbesteuerung (§25a UStG) use CategoryExempt
	// with TaxRate 0 and ExemptionCode VATExSecondHandGoods.
	TaxCategory TaxCategory
	TaxRate     decimal.Decimal

	// ExemptionCode (BT-121) and ExemptionReason (BT-120) are required for
	// non-standard categories (E, AE, K). If ExemptionReason is empty, the
	// default German text for a known ExemptionCode is used.
	ExemptionCode   VATExCode
	ExemptionReason string

	// NetAmount overrides the line net amount (BT-131), which otherwise is
	// Quantity * UnitPrice. Set it when the caller rounds or distributes line
	// nets itself, e.g. on gross-priced documents.
	NetAmount *decimal.Decimal

	// TaxAmount is the line's precomputed VAT. When every line of a VAT
	// breakdown group (BG-23) carries one and the group has no document-level
	// allowance or charge, the group's VAT (BT-117) is their sum instead of
	// basis * rate. This lets gross-priced documents (VAT taken out of the
	// gross sum) carry the VAT their PDF shows. Serialize rejects a group VAT
	// that differs from basis * rate by more than 0.01, or any VAT on a 0 %
	// group, since the XML would fail BR-CO-17.
	TaxAmount *decimal.Decimal
}
