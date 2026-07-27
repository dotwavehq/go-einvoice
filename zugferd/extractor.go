package zugferd

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// maxXMLSize caps the decompressed attachment size. Real e-invoice XMLs
// are a few hundred KB at most; the cap keeps a maliciously compressed
// attachment (decompression bomb) from exhausting memory.
const maxXMLSize = 20 << 20

// ExtractXML returns the embedded e-invoice XML from a ZUGFeRD / Factur-X /
// XRechnung PDF. It looks for the well-known attachment names first and
// falls back to the only XML attachment if exactly one exists. Returns an
// error when the PDF carries no invoice XML — callers use that as the
// signal to fall back to OCR/AI extraction.
func ExtractXML(pdf []byte) ([]byte, error) {
	atts, err := api.ExtractAttachmentsRaw(bytes.NewReader(pdf), "", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("zugferd: read attachments: %w", err)
	}

	wellKnown := map[string]bool{
		"factur-x.xml":        true,
		"zugferd-invoice.xml": true,
		"xrechnung.xml":       true,
		"cii.xml":             true, // Peppol-style name seen in the wild
	}
	var xmlAtts []int
	for i, att := range atts {
		name := strings.ToLower(att.FileName)
		if wellKnown[name] {
			return readCapped(att.Reader)
		}
		if strings.HasSuffix(name, ".xml") {
			xmlAtts = append(xmlAtts, i)
		}
	}
	if len(xmlAtts) == 1 {
		return readCapped(atts[xmlAtts[0]].Reader)
	}
	return nil, fmt.Errorf("zugferd: no invoice XML attachment found (%d attachments)", len(atts))
}

func readCapped(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxXMLSize+1))
	if err != nil {
		return nil, fmt.Errorf("zugferd: read attachment: %w", err)
	}
	if len(data) > maxXMLSize {
		return nil, fmt.Errorf("zugferd: attachment exceeds %d bytes", maxXMLSize)
	}
	return data, nil
}
