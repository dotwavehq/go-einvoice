package einvoice

import "github.com/shopspring/decimal"

// AllowanceCharge is a document-level allowance (discount) or charge (BG-20 /
// BG-21). Its TaxCategory and TaxRate must match the goods they apply to so the
// amount lands in the right VAT breakdown group. Reason is required by XRechnung.
type AllowanceCharge struct {
	Amount      decimal.Decimal
	IsCharge    bool // false = allowance (reduces total), true = charge (adds)
	TaxCategory TaxCategory
	TaxRate     decimal.Decimal
	Reason      string
}

// TaxCategory is an EN 16931 / UNTDID 5305 VAT category code (BT-118 / BT-151).
type TaxCategory string

const (
	CategoryStandard       TaxCategory = "S"  // Standard rate (19% / 7%)
	CategoryZero           TaxCategory = "Z"  // Zero rated goods
	CategoryExempt         TaxCategory = "E"  // Exempt from VAT (incl. §25a margin scheme)
	CategoryReverseCharge  TaxCategory = "AE" // VAT reverse charge (§13b UStG)
	CategoryIntraCommunity TaxCategory = "K"  // Intra-community supply
	CategoryExport         TaxCategory = "G"  // Free export item, tax not charged
	CategoryOutOfScope     TaxCategory = "O"  // Services outside scope of tax
)

// VATExCode is an EN 16931 VAT exemption reason code (BT-121), from the CEF
// VATEX code list. Codes F/I/J cover the German Differenzbesteuerung variants.
type VATExCode string

const (
	VATExSecondHandGoods VATExCode = "VATEX-EU-F"  // Gebrauchtgegenstände (e.g. used cars)
	VATExWorksOfArt      VATExCode = "VATEX-EU-I"  // Kunstgegenstände
	VATExCollectors      VATExCode = "VATEX-EU-J"  // Sammlungsstücke und Antiquitäten
	VATExReverseCharge   VATExCode = "VATEX-EU-AE" // Steuerschuldnerschaft des Leistungsempfängers
	VATExIntraCommunity  VATExCode = "VATEX-EU-IC" // Innergemeinschaftliche Lieferung
)

// DefaultReason returns the standard German exemption text (BT-120) mandated for
// a known VATEX code, or "" if the code is unknown.
func (c VATExCode) DefaultReason() string {
	switch c {
	case VATExSecondHandGoods:
		return "Gebrauchtgegenstände/Sonderregelung"
	case VATExWorksOfArt:
		return "Kunstgegenstände/Sonderregelung"
	case VATExCollectors:
		return "Sammlungsstücke und Antiquitäten/Sonderregelung"
	case VATExReverseCharge:
		return "Steuerschuldnerschaft des Leistungsempfängers"
	case VATExIntraCommunity:
		return "Steuerfreie innergemeinschaftliche Lieferung"
	default:
		return ""
	}
}
