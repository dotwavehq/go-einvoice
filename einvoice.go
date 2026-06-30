package einvoice

import (
	"time"

	"github.com/shopspring/decimal"
)

type Invoice struct {
	Number       string
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
}
