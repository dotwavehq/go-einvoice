package zugferd

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/filter"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// ConformanceLevelEN16931 is the Factur-X / ZUGFeRD profile that matches the
// CII XML produced by the cii package. It is written into the XMP metadata.
const ConformanceLevelEN16931 = "EN 16931"

// facturXNamespace is the Factur-X XMP extension-schema namespace (BG-level
// identification of the embedded invoice).
const facturXNamespace = "urn:factur-x:pdfa:CrossIndustryDocument:invoice:1p0#"

// srgbICC is the HP/Microsoft sRGB IEC61966-2.1 profile (1998), the canonical
// freely redistributable sRGB profile. PDF/A requires a device-independent
// colour space, declared here via the output intent.
//
//go:embed srgb.icc
var srgbICC []byte

// applyPDFA3 turns ctx into a ZUGFeRD/Factur-X PDF/A-3 document: it embeds the
// invoice XML as an associated file and adds the metadata + output intent that
// PDF/A-3 mandates.
//
// It adds the required structure; it does not repair non-conformant content.
// The caller must supply a source PDF that already embeds its fonts and uses
// device RGB (a Chromium / headless-Chrome "print to PDF" satisfies this).
func applyPDFA3(ctx *model.Context, xml []byte, conformanceLevel string, now time.Time) error {
	xt := ctx.XRefTable

	fileSpecRef, err := embedFacturX(xt, xml, now)
	if err != nil {
		return fmt.Errorf("embed factur-x xml: %w", err)
	}

	cat, err := xt.Catalog()
	if err != nil {
		return err
	}

	// PDF/A-3 associated file: links the embedded XML to the document.
	cat["AF"] = types.Array{*fileSpecRef}

	metaRef, err := addMetadataStream(xt, buildXMP(conformanceLevel, now))
	if err != nil {
		return fmt.Errorf("add xmp metadata: %w", err)
	}
	cat["Metadata"] = *metaRef

	outputIntent, err := buildOutputIntent(xt)
	if err != nil {
		return fmt.Errorf("build output intent: %w", err)
	}
	cat["OutputIntents"] = types.Array{outputIntent}

	return nil
}

// embedFacturX adds the XML as an EmbeddedFile with the AFRelationship and
// stream parameters PDF/A-3 requires, registers it in the EmbeddedFiles name
// tree (so any viewer lists it as an attachment), and returns the file-spec
// reference for the catalog /AF entry.
func embedFacturX(xt *model.XRefTable, xml []byte, now time.Time) (*types.IndirectRef, error) {
	ef := &types.StreamDict{
		Dict: types.Dict{
			"Type":    types.Name("EmbeddedFile"),
			"Subtype": types.Name("text/xml"), // written as /text#2Fxml
			"Params": types.Dict{
				"ModDate": types.StringLiteral(types.DateString(now)),
				"Size":    types.Integer(len(xml)),
			},
		},
		Content:        xml,
		FilterPipeline: []types.PDFFilter{{Name: filter.Flate}},
	}
	ef.InsertName("Filter", filter.Flate)
	if err := ef.Encode(); err != nil {
		return nil, err
	}
	efRef, err := xt.IndRefForNewObject(*ef)
	if err != nil {
		return nil, err
	}

	fileSpec := types.Dict{
		"Type":           types.Name("Filespec"),
		"F":              types.StringLiteral(FacturXFileName),
		"UF":             types.StringLiteral(FacturXFileName),
		"Desc":           types.StringLiteral("Factur-X/ZUGFeRD electronic invoice"),
		"AFRelationship": types.Name("Alternative"),
		"EF": types.Dict{
			"F":  *efRef,
			"UF": *efRef,
		},
	}
	fileSpecRef, err := xt.IndRefForNewObject(fileSpec)
	if err != nil {
		return nil, err
	}

	if err := xt.LocateNameTree("EmbeddedFiles", true); err != nil {
		return nil, err
	}
	m := model.NameMap{FacturXFileName: []types.Dict{fileSpec}}
	if err := xt.Names["EmbeddedFiles"].Add(xt, FacturXFileName, *fileSpecRef, m, []string{"F", "UF"}); err != nil {
		return nil, err
	}

	return fileSpecRef, nil
}

// addMetadataStream stores the XMP packet as the document-level /Metadata.
// PDF/A requires this stream to be uncompressed, so no filter is applied.
func addMetadataStream(xt *model.XRefTable, xmp string) (*types.IndirectRef, error) {
	sd := &types.StreamDict{
		Dict: types.Dict{
			"Type":    types.Name("Metadata"),
			"Subtype": types.Name("XML"),
		},
		Content: []byte(xmp),
	}
	if err := sd.Encode(); err != nil {
		return nil, err
	}
	return xt.IndRefForNewObject(*sd)
}

// buildOutputIntent embeds the sRGB profile and returns the GTS_PDFA1 output
// intent dictionary referencing it.
func buildOutputIntent(xt *model.XRefTable) (types.Dict, error) {
	icc := &types.StreamDict{
		Dict:           types.Dict{"N": types.Integer(3)},
		Content:        srgbICC,
		FilterPipeline: []types.PDFFilter{{Name: filter.Flate}},
	}
	icc.InsertName("Filter", filter.Flate)
	if err := icc.Encode(); err != nil {
		return nil, err
	}
	iccRef, err := xt.IndRefForNewObject(*icc)
	if err != nil {
		return nil, err
	}

	return types.Dict{
		"Type":                      types.Name("OutputIntent"),
		"S":                         types.Name("GTS_PDFA1"),
		"OutputConditionIdentifier": types.StringLiteral("sRGB IEC61966-2.1"),
		"Info":                      types.StringLiteral("sRGB IEC61966-2.1"),
		"DestOutputProfile":         *iccRef,
	}, nil
}

// buildXMP renders the XMP packet: the PDF/A identifier (part 3, level B), the
// minimal Dublin Core / XMP basic properties, the Factur-X document fields, and
// the extension-schema description that PDF/A requires for the custom fx
// namespace.
func buildXMP(conformanceLevel string, now time.Time) string {
	ts := now.UTC().Format(time.RFC3339)
	return fmt.Sprintf(`<?xpacket begin="" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
      <pdfaid:part>3</pdfaid:part>
      <pdfaid:conformance>B</pdfaid:conformance>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">
      <dc:format>application/pdf</dc:format>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">
      <xmp:CreatorTool>go-einvoice</xmp:CreatorTool>
      <xmp:CreateDate>%[1]s</xmp:CreateDate>
      <xmp:ModifyDate>%[1]s</xmp:ModifyDate>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:fx="%[2]s">
      <fx:DocumentType>INVOICE</fx:DocumentType>
      <fx:DocumentFileName>%[3]s</fx:DocumentFileName>
      <fx:Version>1.0</fx:Version>
      <fx:ConformanceLevel>%[4]s</fx:ConformanceLevel>
    </rdf:Description>
    <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
      <pdfaExtension:schemas>
        <rdf:Bag>
          <rdf:li rdf:parseType="Resource">
            <pdfaSchema:schema>Factur-X PDFA Extension Schema</pdfaSchema:schema>
            <pdfaSchema:namespaceURI>%[2]s</pdfaSchema:namespaceURI>
            <pdfaSchema:prefix>fx</pdfaSchema:prefix>
            <pdfaSchema:property>
              <rdf:Seq>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>DocumentType</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>INVOICE</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>DocumentFileName</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>name of the embedded XML invoice file</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>Version</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>version of the Factur-X standard</pdfaProperty:description>
                </rdf:li>
                <rdf:li rdf:parseType="Resource">
                  <pdfaProperty:name>ConformanceLevel</pdfaProperty:name>
                  <pdfaProperty:valueType>Text</pdfaProperty:valueType>
                  <pdfaProperty:category>external</pdfaProperty:category>
                  <pdfaProperty:description>conformance level of the embedded invoice</pdfaProperty:description>
                </rdf:li>
              </rdf:Seq>
            </pdfaSchema:property>
          </rdf:li>
        </rdf:Bag>
      </pdfaExtension:schemas>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`, ts, facturXNamespace, FacturXFileName, conformanceLevel)
}
