package model

// InvoiceSerializer defines the contract for converting the domain model into a spec format.
type InvoiceSerializer interface {
	Serialize(invoice *Invoice) ([]byte, error)
}
