package zugferd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// FacturXFileName is the mandatory filename for the XML attachment in ZUGFeRD.
const FacturXFileName = "factur-x.xml"

// EmbedXML attaches the generated XML bytes to an existing PDF file.
// It creates a ZUGFeRD compatible PDF (visual + XML data).
//
// Note: To be fully compliant with standard validation (PDF/A-3),
// the input PDF should ideally already be valid PDF/A.
// This function handles the file attachment structure.
func EmbedXML(inputPDFPath string, xmlContent []byte, outputPDFPath string) error {
	tmpDir, err := os.MkdirTemp("", "go-einvoice")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpXMLPath := filepath.Join(tmpDir, FacturXFileName)
	if err := os.WriteFile(tmpXMLPath, xmlContent, 0644); err != nil {
		return fmt.Errorf("failed to write temp xml: %w", err)
	}

	conf := model.NewDefaultConfiguration()

	err = api.AddAttachmentsFile(
		inputPDFPath,
		outputPDFPath,
		[]string{tmpXMLPath},
		true,
		conf,
	)

	if err != nil {
		return fmt.Errorf("pdfcpu failed to attach file: %w", err)
	}

	return nil
}

// GenerateXMPTemplate returns the RDF metadata required for a valid PDF/A-3 ZUGFeRD file.
func GenerateXMPTemplate() string {
	return `<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:zf="urn:ferd:pdfa:CrossIndustryDocument:invoice:1p0#">
      <zf:DocumentType>INVOICE</zf:DocumentType>
      <zf:DocumentFileName>factur-x.xml</zf:DocumentFileName>
      <zf:Version>1.0</zf:Version>
      <zf:ConformanceLevel>EN 16931</zf:ConformanceLevel>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/"
       xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#"
       xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
      <pdfaExtension:schemas>
        <rdf:Bag>
          <rdf:li rdf:parseType="Resource">
            <pdfaSchema:schema>ZUGFeRD PDFA Extension Schema</pdfaSchema:schema>
            <pdfaSchema:namespaceURI>urn:ferd:pdfa:CrossIndustryDocument:invoice:1p0#</pdfaSchema:namespaceURI>
            <pdfaSchema:prefix>zf</pdfaSchema:prefix>
            <pdfaSchema:property>
              <rdf:Seq>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>name of the embedded XML invoice file</pdfaProperty:description>
                </rdf:li>
              </rdf:Seq>
            </pdfaSchema:property>
          </rdf:li>
        </rdf:Bag>
      </pdfaExtension:schemas>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`
}
