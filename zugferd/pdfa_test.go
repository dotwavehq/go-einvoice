package zugferd

import (
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// TestEmbedXML_ProducesPDFA3 embeds XML into a freshly generated PDF and
// asserts the PDF/A-3 / Factur-X structure is wired up: an associated file
// (AFRelationship=Alternative) reachable from the catalog /AF, an sRGB output
// intent (S=GTS_PDFA1) and an XMP /Metadata stream.
func TestEmbedXML_ProducesPDFA3(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.pdf")
	out := filepath.Join(dir, "out.pdf")

	conf := model.NewDefaultConfiguration()
	blank, err := pdfcpu.CreateContextWithXRefTable(conf, &types.Dim{Width: 595, Height: 842})
	if err != nil {
		t.Fatalf("create blank pdf: %v", err)
	}
	if err := api.WriteContextFile(blank, in); err != nil {
		t.Fatalf("write blank pdf: %v", err)
	}

	xml := []byte(`<?xml version="1.0" encoding="UTF-8"?><rsm:CrossIndustryInvoice/>`)
	if err := EmbedXML(in, xml, out); err != nil {
		t.Fatalf("EmbedXML: %v", err)
	}

	ctx, err := api.ReadContextFile(out)
	if err != nil {
		t.Fatalf("read output pdf: %v", err)
	}
	xt := ctx.XRefTable
	cat, err := xt.Catalog()
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}

	if _, ok := cat["Metadata"]; !ok {
		t.Error("catalog is missing /Metadata (XMP)")
	}

	af, err := xt.DereferenceArray(cat["AF"])
	if err != nil || len(af) != 1 {
		t.Fatalf("catalog /AF: err=%v len=%d, want 1", err, len(af))
	}
	fs, err := xt.DereferenceDict(af[0])
	if err != nil {
		t.Fatalf("dereference file spec: %v", err)
	}
	if rel, _ := fs["AFRelationship"].(types.Name); rel != "Alternative" {
		t.Errorf("AFRelationship = %q, want Alternative", rel)
	}

	oi, err := xt.DereferenceArray(cat["OutputIntents"])
	if err != nil || len(oi) != 1 {
		t.Fatalf("catalog /OutputIntents: err=%v len=%d, want 1", err, len(oi))
	}
	oiDict, err := xt.DereferenceDict(oi[0])
	if err != nil {
		t.Fatalf("dereference output intent: %v", err)
	}
	if s, _ := oiDict["S"].(types.Name); s != "GTS_PDFA1" {
		t.Errorf("OutputIntent /S = %q, want GTS_PDFA1", s)
	}
}
