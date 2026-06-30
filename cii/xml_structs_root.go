package cii

import "encoding/xml"

type CrossIndustryInvoice struct {
	XMLName xml.Name `xml:"rsm:CrossIndustryInvoice"`
	Rsm     string   `xml:"xmlns:rsm,attr"`
	Ram     string   `xml:"xmlns:ram,attr"`
	Udt     string   `xml:"xmlns:udt,attr"`

	Context     ExchangedDocumentContext    `xml:"rsm:ExchangedDocumentContext"`
	Header      ExchangedDocument           `xml:"rsm:ExchangedDocument"`
	Transaction SupplyChainTradeTransaction `xml:"rsm:SupplyChainTradeTransaction"`
}

type ExchangedDocumentContext struct {
	// NEU: Business Process (R001)
	BusinessProcess    IDType `xml:"ram:BusinessProcessSpecifiedDocumentContextParameter>ram:ID,omitempty"`
	GuidelineParameter IDType `xml:"ram:GuidelineSpecifiedDocumentContextParameter>ram:ID"`
}

type ExchangedDocument struct {
	ID           string   `xml:"ram:ID"`
	TypeCode     string   `xml:"ram:TypeCode"`
	IssueDate    DateType `xml:"ram:IssueDateTime>udt:DateTimeString"`
	IncludedNote *Note    `xml:"ram:IncludedNote,omitempty"`
}

type IDType struct {
	Value string `xml:",chardata"`
}

type DateType struct {
	Value  string `xml:",chardata"`
	Format string `xml:"format,attr"`
}

type Note struct {
	Content string `xml:"ram:Content"`
}
