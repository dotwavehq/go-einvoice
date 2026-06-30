// Package einvoice provides the domain model for German E-Invoices (E-Rechnung)
// compliant with EN 16931.
//
// The Invoice type is the entry point. Serialize it to XRechnung 3.0 / CII XML
// with the cii package, and embed that XML into a PDF/A-3 (ZUGFeRD / Factur-X)
// with the zugferd package.
package einvoice
