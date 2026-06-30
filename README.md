# go-einvoice

[![CI](https://github.com/dotwavehq/go-einvoice/actions/workflows/ci.yml/badge.svg)](https://github.com/dotwavehq/go-einvoice/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dotwavehq/go-einvoice.svg)](https://pkg.go.dev/github.com/dotwavehq/go-einvoice)
[![Go Report Card](https://goreportcard.com/badge/github.com/dotwavehq/go-einvoice)](https://goreportcard.com/report/github.com/dotwavehq/go-einvoice)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](LICENSE)

A Go library for generating German E-Invoices (E-Rechnung) compliant with EN 16931.

This library is designed to help developers upgrade their invoicing systems for the upcoming German B2B mandates starting January 1st, 2025. It supports the creation of XRechnung 3.0 (XML) and ZUGFeRD / Factur-X (PDF/A-3 with embedded XML).

## Features

✅ Full support for EN 16931 (UN/CEFACT CII).
✅ Compliant with XRechnung 3.0 (German Standard).
✅ Generates ZUGFeRD / Factur-X files by embedding XML into existing PDFs.

## Installation

```go
go get github.com/dotwavehq/go-einvoice
```

## Library Usage

To use `go-einvoice` in your backend service, you build an `einvoice.Invoice` and hand it to a serializer. The library takes care of the complex XML namespaces and business rules.

Create an Invoice

```go
import (
    einvoice "github.com/dotwavehq/go-einvoice"
    "github.com/dotwavehq/go-einvoice/cii"
    "github.com/dotwavehq/go-einvoice/zugferd"
    "github.com/shopspring/decimal"
)

func main() {
    // Helper to create decimals from string to avoid float precision errors
    toDec := func(s string) decimal.Decimal {
        d, _ := decimal.NewFromString(s)
        return d
    }

    invoice := einvoice.Invoice{
        Number:         "RE-2025-1001",
        IssueDate:      time.Now(),
        DueDate:        time.Now().AddDate(0, 0, 14),
        DeliveryDate:   time.Now(), // BT-72, required under German VAT law
        BuyerReference: "ORDER-12345",
        Currency:       "EUR",
        Note:           "Thank you for your business.",

        Seller: einvoice.Party{
            Name:             "My Software Company GmbH",
            Street:           "Tech Lane 1",
            City:             "Berlin",
            PostalCode:       "10115",
            CountryCode:      "DE",
            VATID:            "DE123456789",
            ElectronicAddress: "billing@mycompany.com", // BT-34, mandatory in XRechnung
            Contact: &einvoice.Contact{
                Name:  "Jane Doe",
                Phone: "+49 30 123456",
                Email: "billing@mycompany.com",
            },
        },

        Buyer: einvoice.Party{
            Name:             "Client Corp AG",
            Street:           "Business Rd 5",
            City:             "Munich",
            PostalCode:       "80331",
            CountryCode:      "DE",
            VATID:            "DE987654321",
            ElectronicAddress: "ap@clientcorp.de", // BT-49, mandatory in XRechnung
        },

        Payment: einvoice.Payment{
            IBAN:             "DE99123456789012345678",
            PaymentMeansCode: "30",
        },

        LineItems: []einvoice.LineItem{
            {
                Description: "Consulting Services",
                Quantity:    toDec("10.0"),
                UnitCode:    "HUR",
                UnitPrice:   toDec("100.00"),
                TaxCategory: einvoice.CategoryStandard,
                TaxRate:     toDec("19.0"),
            },
            {
                Description: "Hosting Fee",
                Quantity:    toDec("1.0"),
                UnitCode:    "C62",
                UnitPrice:   toDec("50.00"),
                TaxCategory: einvoice.CategoryStandard,
                TaxRate:     toDec("19.0"),
            },
        },
        // Totals (net, VAT per rate, grand total) are computed from the lines.
    }

    // Serialize to XML (CII / XRechnung format)
    serializer := cii.NewSerializer()
    xmlBytes, err := serializer.Serialize(&invoice)
    if err != nil {
        log.Fatalf("Failed to generate XML: %v", err)
    }

    // Save as pure XRechnung XML
    os.WriteFile("xrechnung.xml", xmlBytes, 0644)

    // Create ZUGFeRD (Hybrid PDF)
    err = zugferd.EmbedXML("invoice.pdf", xmlBytes, "zugferd.pdf")
    if err != nil {
        log.Fatalf("Failed to embed XML into PDF: %v", err)
    }
    
    log.Println("Successfully generated E-Invoices!")
}
```

### Differenzbesteuerung (§25a UStG)

For margin-scheme goods (used cars, art, antiques) VAT is not shown separately.
Use category `E` with rate `0` and the matching VATEX exemption code; the library
emits the exemption reason (BT-120/BT-121) and the mandatory legal note (BG-1):

```go
einvoice.LineItem{
    Description:   "Gebrauchtwagen VW Golf",
    Quantity:      toDec("1"),
    UnitCode:      "C62",
    UnitPrice:     toDec("10000.00"),
    TaxCategory:   einvoice.CategoryExempt,      // "E"
    TaxRate:       toDec("0"),
    ExemptionCode: einvoice.VATExSecondHandGoods, // VATEX-EU-F → "Gebrauchtgegenstände/Sonderregelung"
}
```

Mixing rates on one invoice (e.g. a margin-scheme car plus a 19% delivery fee) is
supported — each (category, rate) combination becomes its own VAT breakdown group.

Other categories work the same way: `CategoryIntraCommunity` (`K`, tax-free EU
supply — a deliver-to address is added automatically) and `CategoryReverseCharge`
(`AE`, §13b) with their matching `VATEx*` exemption codes.

## CLI Usage

### Generate XML

```go
./einvoice -in invoice.json -out invoice
# Creates invoice.xml
```
### Generate ZUGFeRD PDF

```go
./einvoice -in invoice.json -pdf invoice.pdf -out invoice
# Creates invoice.pdf
```

## Validation

Generated invoices are validated in CI against the official **EN 16931** and
**XRechnung (KoSIT)** schematron rules. To run it locally you need `python3`
with [`saxonche`](https://pypi.org/project/saxonche/) (an XSLT 2.0 engine):

```bash
pip install saxonche
bash scripts/validate.sh   # builds example/invoice.json → CII XML → validates
```

The script downloads and caches the official schematron and fails on any
blocking (`fatal`/`error`) rule violation.

## License

GNU AFFERO GENERAL PUBLIC LICENSE
