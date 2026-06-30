package einvoice_test

import (
	"fmt"

	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/dotwavehq/go-einvoice/cii"
	"github.com/shopspring/decimal"
)

func ExampleInvoice() {
	inv := &einvoice.Invoice{
		Number:   "RE-2025-1001",
		Currency: "EUR",
		Seller:   einvoice.Party{Name: "Seller GmbH", CountryCode: "DE", VATID: "DE123456789"},
		Buyer:    einvoice.Party{Name: "Buyer AG", CountryCode: "DE"},
		LineItems: []einvoice.LineItem{{
			Description: "Consulting",
			Quantity:    decimal.NewFromInt(10),
			UnitCode:    "HUR",
			UnitPrice:   decimal.RequireFromString("100.00"),
			TaxCategory: einvoice.CategoryStandard,
			TaxRate:     decimal.NewFromInt(19),
		}},
	}

	xmlBytes, err := cii.NewSerializer().Serialize(inv)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(xmlBytes) > 0)
	// Output: true
}
