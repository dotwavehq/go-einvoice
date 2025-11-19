package cii

import (
	"encoding/xml"
	"strconv"

	"github.com/dotwavehq/go-einvoice/pkg/model"
)

type CIISerializer struct{}

func NewSerializer() *CIISerializer {
	return &CIISerializer{}
}

func (s *CIISerializer) Serialize(inv *model.Invoice) ([]byte, error) {
	buyerRef := inv.BuyerReference
	if buyerRef == "" {
		buyerRef = "NOT_PROVIDED"
	}

	createElectronicAddress := func(vatID string) *UniversalCommunication {
		if vatID == "" {
			return nil
		}
		return &UniversalCommunication{
			URIID: &IDWithScheme{
				Value:    vatID,
				SchemeID: "EM",
			},
		}
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
			ElectronicURI: createElectronicAddress(inv.Seller.VATID),
		},
		Buyer: TradeParty{
			Name: inv.Buyer.Name,
			PostalAddress: TradeAddress{
				Postcode:  inv.Buyer.PostalCode,
				LineOne:   inv.Buyer.Street,
				CityName:  inv.Buyer.City,
				CountryID: inv.Buyer.CountryCode,
			},
			ElectronicURI: createElectronicAddress(inv.Buyer.VATID),
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
					CategoryCode:          "S",
					RateApplicablePercent: fmtAmount2(item.TaxRate),
				},
				Summation: LineMonetarySummation{
					LineTotalAmount: Amount{Value: fmtAmount2(lineTotal)},
				},
			},
		}
		xmlLines = append(xmlLines, xmlLine)
	}

	tradeTax := TradeTax{
		CalculatedAmount:      &Amount{Value: fmtAmount2(inv.TaxTotal)},
		TypeCode:              "VAT",
		BasisAmount:           &Amount{Value: fmtAmount2(inv.GrandTotal.Sub(inv.TaxTotal))},
		CategoryCode:          "S",
		RateApplicablePercent: "19.00",
	}

	settlement := Settlement{
		InvoiceCurrency:      inv.Currency,
		ApplicableTradeTaxes: []TradeTax{tradeTax},
		MonetarySummation: MonetarySummation{
			LineTotalAmount:      Amount{Value: fmtAmount2(inv.GrandTotal.Sub(inv.TaxTotal))},
			TaxBasisTotalAmount:  Amount{Value: fmtAmount2(inv.GrandTotal.Sub(inv.TaxTotal))},
			TaxTotalAmount:       AmountWithCurrency{Value: fmtAmount2(inv.TaxTotal), Currency: inv.Currency},
			GrandTotalAmount:     Amount{Value: fmtAmount2(inv.GrandTotal)},
			DuePayableAmount:     Amount{Value: fmtAmount2(inv.GrandTotal)},
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
			IncludedNote: &Note{Content: inv.Note},
		},
		Transaction: SupplyChainTradeTransaction{
			IncludedLineItems: xmlLines,
			Agreement:         agreement,
			Settlement:        settlement,
		},
	}

	output, err := xml.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}
