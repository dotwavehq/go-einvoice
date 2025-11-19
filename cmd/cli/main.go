package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/dotwavehq/go-einvoice/pkg/cii"
	"github.com/dotwavehq/go-einvoice/pkg/model"
	"github.com/dotwavehq/go-einvoice/pkg/zugferd"
)

func main() {
	inputJSON := flag.String("in", "invoice.json", "Path to input JSON data")
	inputPDF := flag.String("pdf", "", "Optional: Path to input PDF (visual invoice). If set, output will be ZUGFeRD PDF.")
	output := flag.String("out", "invoice_out", "Output filename prefix (without extension)")
	flag.Parse()

	fmt.Printf("Reading JSON from %s...\n", *inputJSON)
	data, err := os.ReadFile(*inputJSON)
	if err != nil {
		panic(fmt.Sprintf("Could not read input: %v", err))
	}

	var invoice model.Invoice
	if err := json.Unmarshal(data, &invoice); err != nil {
		panic(fmt.Sprintf("Invalid JSON format: %v", err))
	}

	fmt.Println("Generating XML ...")
	serializer := cii.NewSerializer()
	xmlBytes, err := serializer.Serialize(&invoice)
	if err != nil {
		panic(fmt.Sprintf("Generation failed: %v", err))
	}

	if *inputPDF != "" {
		// ZUGFeRD
		outPDF := *output + ".pdf"
		fmt.Printf("Combining '%s' with XML into ZUGFeRD PDF: %s\n", *inputPDF, outPDF)

		err := zugferd.EmbedXML(*inputPDF, xmlBytes, outPDF)
		if err != nil {
			panic(fmt.Sprintf("ZUGFeRD embedding failed: %v", err))
		}
		fmt.Println("✅ Success! ZUGFeRD PDF created.")

	} else {
		// Pure XML
		outXML := *output + ".xml"
		fmt.Printf("Saving pure XRechnung/CII XML: %s\n", outXML)

		if err := os.WriteFile(outXML, xmlBytes, 0644); err != nil {
			panic(err)
		}
		fmt.Println("✅ Success! XML created.")
	}
}
