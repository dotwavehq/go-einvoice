package einvoice

import (
	"time"

	"github.com/shopspring/decimal"
)

type Invoice struct {
	Number         string
	IssueDate      time.Time
	DueDate        time.Time
	Currency       string
	Note           string
	BuyerReference string

	Seller    Party
	Buyer     Party
	Payment   Payment
	LineItems []LineItem
}

type Party struct {
	Name        string
	Street      string
	City        string
	PostalCode  string
	CountryCode string
	VATID       string
	Contact     *Contact
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
