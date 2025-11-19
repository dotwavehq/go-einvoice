package cii

type SupplyChainTradeTransaction struct {
	IncludedLineItems []LineItem `xml:"ram:IncludedSupplyChainTradeLineItem"`
	Agreement         Agreement  `xml:"ram:ApplicableHeaderTradeAgreement"`
	Delivery          Delivery   `xml:"ram:ApplicableHeaderTradeDelivery,omitempty"`
	Settlement        Settlement `xml:"ram:ApplicableHeaderTradeSettlement"`
}

type Agreement struct {
	BuyerReference string     `xml:"ram:BuyerReference,omitempty"`
	Seller         TradeParty `xml:"ram:SellerTradeParty"`
	Buyer          TradeParty `xml:"ram:BuyerTradeParty"`
}

type TradeParty struct {
	ID              []TradeID               `xml:"ram:ID,omitempty"`
	GlobalID        *TradeID                `xml:"ram:GlobalID,omitempty"`
	Name            string                  `xml:"ram:Name"`
	Contact         *DefinedTradeContact    `xml:"ram:DefinedTradeContact,omitempty"`
	PostalAddress   TradeAddress            `xml:"ram:PostalTradeAddress"`
	ElectronicURI   *UniversalCommunication `xml:"ram:URIUniversalCommunication,omitempty"`
	TaxRegistration *TaxRegistration        `xml:"ram:SpecifiedTaxRegistration,omitempty"`
}

type UniversalCommunication struct {
	URIID *IDWithScheme `xml:"ram:URIID,omitempty"`
}

type IDWithScheme struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr,omitempty"`
}

type TradeID struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr,omitempty"`
}

type DefinedTradeContact struct {
	PersonName string `xml:"ram:PersonName,omitempty"`
	Telephone  *Phone `xml:"ram:TelephoneUniversalCommunication,omitempty"`
	Email      *Email `xml:"ram:EmailURIUniversalCommunication,omitempty"`
}

type Phone struct {
	CompleteNumber string `xml:"ram:CompleteNumber"`
}

type Email struct {
	URIID string `xml:"ram:URIID"`
}

type TradeAddress struct {
	Postcode  string `xml:"ram:PostcodeCode,omitempty"`
	LineOne   string `xml:"ram:LineOne,omitempty"`
	CityName  string `xml:"ram:CityName,omitempty"`
	CountryID string `xml:"ram:CountryID"`
}

type TaxRegistration struct {
	ID TaxID `xml:"ram:ID"`
}

type TaxID struct {
	Value    string `xml:",chardata"`
	SchemeID string `xml:"schemeID,attr"`
}

type Delivery struct {
	Occurrence *Occurrence `xml:"ram:ActualDeliverySupplyChainEvent,omitempty"`
}

type Occurrence struct {
	OccurrenceDate DateType `xml:"ram:OccurrenceDateTime>udt:DateTimeString"`
}

type Settlement struct {
	InvoiceCurrency      string            `xml:"ram:InvoiceCurrencyCode"`
	PaymentMeans         *PaymentMeans     `xml:"ram:SpecifiedTradeSettlementPaymentMeans,omitempty"`
	ApplicableTradeTaxes []TradeTax        `xml:"ram:ApplicableTradeTax"`
	BillingPeriod        *BillingPeriod    `xml:"ram:BillingSpecifiedPeriod,omitempty"`
	PaymentTerms         *PaymentTerms     `xml:"ram:SpecifiedTradePaymentTerms,omitempty"`
	MonetarySummation    MonetarySummation `xml:"ram:SpecifiedTradeSettlementHeaderMonetarySummation"`
}

type PaymentTerms struct {
	DueDate *DateType `xml:"ram:DueDateDateTime>udt:DateTimeString,omitempty"`
}

type PaymentMeans struct {
	TypeCode     string        `xml:"ram:TypeCode"`
	PayeeAccount *PayeeAccount `xml:"ram:PayeePartyCreditorFinancialAccount,omitempty"`
}

type PayeeAccount struct {
	IBANID string `xml:"ram:IBANID"`
}

type TradeTax struct {
	CalculatedAmount      *Amount `xml:"ram:CalculatedAmount,omitempty"`
	TypeCode              string  `xml:"ram:TypeCode"`
	BasisAmount           *Amount `xml:"ram:BasisAmount,omitempty"`
	CategoryCode          string  `xml:"ram:CategoryCode"`
	RateApplicablePercent string  `xml:"ram:RateApplicablePercent"`
}

type BillingPeriod struct {
	Start DateType `xml:"ram:StartDateTime>udt:DateTimeString,omitempty"`
	End   DateType `xml:"ram:EndDateTime>udt:DateTimeString,omitempty"`
}

type MonetarySummation struct {
	LineTotalAmount      Amount             `xml:"ram:LineTotalAmount"`
	ChargeTotalAmount    Amount             `xml:"ram:ChargeTotalAmount"`
	AllowanceTotalAmount Amount             `xml:"ram:AllowanceTotalAmount"`
	TaxBasisTotalAmount  Amount             `xml:"ram:TaxBasisTotalAmount"`
	TaxTotalAmount       AmountWithCurrency `xml:"ram:TaxTotalAmount"`
	GrandTotalAmount     Amount             `xml:"ram:GrandTotalAmount"`
	DuePayableAmount     Amount             `xml:"ram:DuePayableAmount"`
}

type Amount struct {
	Value string `xml:",chardata"`
}

type AmountWithCurrency struct {
	Value    string `xml:",chardata"`
	Currency string `xml:"currencyID,attr"`
}

type LineItem struct {
	DocumentLineDocument LineDocument   `xml:"ram:AssociatedDocumentLineDocument"`
	Product              TradeProduct   `xml:"ram:SpecifiedTradeProduct"`
	Agreement            LineAgreement  `xml:"ram:SpecifiedLineTradeAgreement"`
	Delivery             LineDelivery   `xml:"ram:SpecifiedLineTradeDelivery"`
	Settlement           LineSettlement `xml:"ram:SpecifiedLineTradeSettlement"`
}

type LineDocument struct {
	LineID string `xml:"ram:LineID"`
}

type TradeProduct struct {
	Name string `xml:"ram:Name"`
}

type LineAgreement struct {
	NetPrice Price `xml:"ram:NetPriceProductTradePrice"`
}

type Price struct {
	ChargeAmount string `xml:"ram:ChargeAmount"`
}

type LineDelivery struct {
	BilledQuantity Quantity `xml:"ram:BilledQuantity"`
}

type Quantity struct {
	Value    string `xml:",chardata"`
	UnitCode string `xml:"unitCode,attr"`
}

type LineSettlement struct {
	Tax       TradeTax              `xml:"ram:ApplicableTradeTax"`
	Summation LineMonetarySummation `xml:"ram:SpecifiedTradeSettlementLineMonetarySummation"`
}

type LineMonetarySummation struct {
	LineTotalAmount Amount `xml:"ram:LineTotalAmount"`
}
