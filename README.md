# go-einvoice

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

To use `go-einvoice` in your backend service, you interact primarily with the model package. The library takes care of the complex XML namespaces and business rules.

Create an Invoice

```go
func main() {
    // Helper to create decimals from string to avoid float precision errors
    toDec := func(s string) decimal.Decimal {
        d, _ := decimal.NewFromString(s)
        return d
    }

    invoice := model.Invoice{
        Number:         "RE-2025-1001",
        IssueDate:      time.Now(),
        DueDate:        time.Now().AddDate(0, 0, 14),
        BuyerReference: "ORDER-12345",
        Currency:       "EUR",
        Note:           "Thank you for your business.",

        Seller: model.Party{
            Name:        "My Software Company GmbH",
            Street:      "Tech Lane 1",
            City:        "Berlin",
            PostalCode:  "10115",
            CountryCode: "DE", 
            VATID:       "DE123456789",
            Contact: &model.Contact{
                Name:  "Jane Doe",
                Phone: "+49 30 123456",
                Email: "billing@mycompany.com",
            },
        },

        Buyer: model.Party{
            Name:        "Client Corp AG",
            Street:      "Business Rd 5",
            City:        "Munich",
            PostalCode:  "80331",
            CountryCode: "DE",
            VATID:       "DE987654321",
        },

        Payment: model.Payment{
            IBAN:             "DE99123456789012345678",
            PaymentMeansCode: "30",
        },

        LineItems: []model.LineItem{
            {
                Description: "Consulting Services",
                Quantity:    toDec("10.0"),
                UnitCode:    "HUR",
                UnitPrice:   toDec("100.00"),
                TaxRate:     toDec("19.0"),
            },
            {
                Description: "Hosting Fee",
                Quantity:    toDec("1.0"),
                UnitCode:    "C62",
                UnitPrice:   toDec("50.00"),
                TaxRate:     toDec("19.0"),
            },
        },
        
        TaxTotal:   toDec("199.50"),
        GrandTotal: toDec("1249.50"),
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

## CLI Usage

### Generate XML

```go
./go-einvoice -in invoice.json -out invoice
# Creates invoice.xml
```
### Generate ZUGFeRD PDF

```go
./go-einvoice -in invoice.json -pdf invoice.pdf -out invoice
# Creates invoice.pdf
```

## License

GNU AFFERO GENERAL PUBLIC LICENSE
