package model

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

	TaxTotal   decimal.Decimal
	GrandTotal decimal.Decimal
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
	TaxRate     decimal.Decimal
}
