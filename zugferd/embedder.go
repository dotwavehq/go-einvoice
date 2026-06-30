package zugferd

import (
	"fmt"
	"os"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FacturXFileName is the mandatory filename for the XML attachment in
// ZUGFeRD / Factur-X.
const FacturXFileName = "factur-x.xml"

// EmbedXML produces a ZUGFeRD / Factur-X PDF/A-3 by embedding the invoice XML
// into an existing PDF: it attaches the XML as the associated file
// factur-x.xml and writes the PDF/A identifier, the Factur-X document
// metadata, and an sRGB output intent.
//
// The XML stays extractable as an ordinary PDF attachment, so a recipient can
// process either the visual PDF or the structured data.
//
// EmbedXML adds the PDF/A-3 structure but does not repair the visual content:
// the input PDF must already embed its fonts and use device RGB. A PDF
// rendered by headless Chromium ("print to PDF") satisfies this. To confirm
// full conformance, validate the output with veraPDF.
func EmbedXML(inputPDFPath string, xmlContent []byte, outputPDFPath string) error {
	in, err := os.Open(inputPDFPath)
	if err != nil {
		return fmt.Errorf("open input pdf: %w", err)
	}
	defer func() { _ = in.Close() }()

	conf := model.NewDefaultConfiguration()
	conf.Cmd = model.ADDATTACHMENTS

	ctx, err := api.ReadValidateAndOptimize(in, conf)
	if err != nil {
		return fmt.Errorf("read input pdf: %w", err)
	}

	if err := applyPDFA3(ctx, xmlContent, ConformanceLevelEN16931, time.Now()); err != nil {
		return fmt.Errorf("apply pdf/a-3: %w", err)
	}

	out, err := os.Create(outputPDFPath)
	if err != nil {
		return fmt.Errorf("create output pdf: %w", err)
	}
	defer func() { _ = out.Close() }()

	if err := api.Write(ctx, out, conf); err != nil {
		return fmt.Errorf("write output pdf: %w", err)
	}

	return nil
}
