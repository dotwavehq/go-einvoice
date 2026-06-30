package cii

import (
	"encoding/xml"
	"strconv"

	einvoice "github.com/dotwavehq/go-einvoice"
	"github.com/shopspring/decimal"
)

type CIISerializer struct{}

func NewSerializer() *CIISerializer {
	return &CIISerializer{}
}

func (s *CIISerializer) Serialize(inv *einvoice.Invoice) ([]byte, error) {
	buyerRef := inv.BuyerReference
	if buyerRef == "" {
		buyerRef = "NOT_PROVIDED"
	}

	agreement := Agreement{
		BuyerReference: buyerRef,
		Seller: TradeParty{
			Name: inv.Seller.Name,
			PostalAddress: TradeAddress{
				Postcode:  inv.Seller.PostalCode,
				LineOne:   inv.Seller.Street,
				CityName:  inv.Seller.City,
				CountryID: inv.Seller.CountryCode,
			},
			TaxRegistration: &TaxRegistration{
				ID: TaxID{Value: inv.Seller.VATID, SchemeID: "VA"},
			},
			ElectronicURI: electronicAddress(inv.Seller),
		},
		Buyer: TradeParty{
			Name: inv.Buyer.Name,
			PostalAddress: TradeAddress{
				Postcode:  inv.Buyer.PostalCode,
				LineOne:   inv.Buyer.Street,
				CityName:  inv.Buyer.City,
				CountryID: inv.Buyer.CountryCode,
			},
			ElectronicURI: electronicAddress(inv.Buyer),
		},
	}

	if inv.Seller.Contact != nil {
		agreement.Seller.Contact = &DefinedTradeContact{
			PersonName: inv.Seller.Contact.Name,
			Telephone:  &Phone{CompleteNumber: inv.Seller.Contact.Phone},
			Email:      &Email{URIID: inv.Seller.Contact.Email},
		}
	}

	if inv.Buyer.VATID != "" {
		agreement.Buyer.TaxRegistration = &TaxRegistration{
			ID: TaxID{Value: inv.Buyer.VATID, SchemeID: "VA"},
		}
	}

	var xmlLines []LineItem
	for i, item := range inv.LineItems {
		lineTotal := item.Quantity.Mul(item.UnitPrice)

		xmlLine := LineItem{
			DocumentLineDocument: LineDocument{
				LineID: strconv.Itoa(i + 1),
			},
			Product: TradeProduct{
				Name: item.Description,
			},
			Agreement: LineAgreement{
				NetPrice: Price{
					ChargeAmount: fmtAmount2(item.UnitPrice),
				},
			},
			Delivery: LineDelivery{
				BilledQuantity: Quantity{
					Value:    fmtQty(item.Quantity),
					UnitCode: item.UnitCode,
				},
			},
			Settlement: LineSettlement{
				Tax: TradeTax{
					TypeCode:              "VAT",
					CategoryCode:          categoryOf(item),
					RateApplicablePercent: fmtAmount2(item.TaxRate),
				},
				Summation: LineMonetarySummation{
					LineTotalAmount: Amount{Value: fmtAmount2(lineTotal)},
				},
			},
		}
		xmlLines = append(xmlLines, xmlLine)
	}

	// One VAT breakdown (BG-23) per (category, rate); totals are computed from
	// the lines so they always reconcile (BR-CO-10..17).
	groups := taxBreakdown(inv)
	lineTotal, taxTotal := decimal.Zero, decimal.Zero
	tradeTaxes := make([]TradeTax, 0, len(groups))
	for _, g := range groups {
		lineTotal = lineTotal.Add(g.basis)
		taxTotal = taxTotal.Add(g.tax)
		tradeTaxes = append(tradeTaxes, TradeTax{
			CalculatedAmount:      &Amount{Value: fmtAmount2(g.tax)},
			TypeCode:              "VAT",
			ExemptionReason:       g.exReason,
			BasisAmount:           &Amount{Value: fmtAmount2(g.basis)},
			CategoryCode:          g.category,
			ExemptionReasonCode:   g.exCode,
			RateApplicablePercent: fmtAmount2(g.rate),
		})
	}
	grandTotal := lineTotal.Add(taxTotal)

	settlement := Settlement{
		InvoiceCurrency:      inv.Currency,
		ApplicableTradeTaxes: tradeTaxes,
		MonetarySummation: MonetarySummation{
			LineTotalAmount:      Amount{Value: fmtAmount2(lineTotal)},
			TaxBasisTotalAmount:  Amount{Value: fmtAmount2(lineTotal)},
			TaxTotalAmount:       AmountWithCurrency{Value: fmtAmount2(taxTotal), Currency: inv.Currency},
			GrandTotalAmount:     Amount{Value: fmtAmount2(grandTotal)},
			DuePayableAmount:     Amount{Value: fmtAmount2(grandTotal)},
			ChargeTotalAmount:    Amount{Value: "0.00"},
			AllowanceTotalAmount: Amount{Value: "0.00"},
		},
	}

	if inv.Payment.IBAN != "" {
		settlement.PaymentMeans = &PaymentMeans{
			TypeCode: "30",
			PayeeAccount: &PayeeAccount{
				IBANID: inv.Payment.IBAN,
			},
		}
		if inv.Payment.PaymentMeansCode != "" {
			settlement.PaymentMeans.TypeCode = inv.Payment.PaymentMeansCode
		}
	}

	if !inv.DueDate.IsZero() {
		settlement.PaymentTerms = &PaymentTerms{
			DueDate: &DateType{
				Value:  inv.DueDate.Format("20060102"),
				Format: DateFormatYYYYMMDD,
			},
		}
	} else {
		settlement.PaymentTerms = &PaymentTerms{
			DueDate: &DateType{
				Value:  inv.IssueDate.Format("20060102"),
				Format: DateFormatYYYYMMDD,
			},
		}
	}

	// BG-1 notes: the free-text note plus the mandatory legal wording for each
	// exemption reason present (e.g. §25a "Gebrauchtgegenstände/Sonderregelung").
	var notes []Note
	if inv.Note != "" {
		notes = append(notes, Note{Content: inv.Note})
	}
	seen := map[string]bool{}
	for _, g := range groups {
		if g.exReason != "" && !seen[g.exReason] {
			seen[g.exReason] = true
			notes = append(notes, Note{Content: g.exReason})
		}
	}

	// Deliver-to address (BG-15) is mandatory for intra-community supply
	// (BR-IC-12). It defaults to the buyer's address; DeliveryCountryCode
	// overrides only the country (BT-80). XRechnung also requires city (BT-77)
	// and post code (BT-78) once the address is present.
	var delivery Delivery
	if inv.DeliveryCountryCode != "" || hasCategory(inv, einvoice.CategoryIntraCommunity) {
		country := inv.DeliveryCountryCode
		if country == "" {
			country = inv.Buyer.CountryCode
		}
		delivery.ShipTo = &TradeParty{
			Name: inv.Buyer.Name,
			PostalAddress: TradeAddress{
				Postcode:  inv.Buyer.PostalCode,
				LineOne:   inv.Buyer.Street,
				CityName:  inv.Buyer.City,
				CountryID: country,
			},
		}
	}
	if !inv.DeliveryDate.IsZero() {
		delivery.Occurrence = &Occurrence{
			OccurrenceDate: DateType{
				Value:  inv.DeliveryDate.Format("20060102"),
				Format: DateFormatYYYYMMDD,
			},
		}
	}

	invoice := CrossIndustryInvoice{
		Rsm: Rsmns,
		Ram: Ramns,
		Udt: Udtns,
		Context: ExchangedDocumentContext{
			BusinessProcess:    IDType{Value: BusinessProcessType},
			GuidelineParameter: IDType{Value: ProfileXRechnung3},
		},
		Header: ExchangedDocument{
			ID:       inv.Number,
			TypeCode: TypeCodeCommercialInvoice,
			IssueDate: DateType{
				Value:  inv.IssueDate.Format("20060102"),
				Format: DateFormatYYYYMMDD,
			},
			IncludedNote: notes,
		},
		Transaction: SupplyChainTradeTransaction{
			IncludedLineItems: xmlLines,
			Agreement:         agreement,
			Delivery:          delivery,
			Settlement:        settlement,
		},
	}

	output, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}
