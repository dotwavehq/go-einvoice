package zugferd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

func TestExtractXMLRoundTrip(t *testing.T) {
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

	xml := []byte(`<?xml version="1.0"?><rsm:CrossIndustryInvoice/>`)
	if err := EmbedXML(in, xml, out); err != nil {
		t.Fatalf("EmbedXML: %v", err)
	}

	pdf, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	got, err := ExtractXML(pdf)
	if err != nil {
		t.Fatalf("ExtractXML: %v", err)
	}
	if !bytes.Equal(got, xml) {
		t.Errorf("extracted = %q, want %q", got, xml)
	}

	// A plain PDF without attachments must signal "no e-invoice".
	plain, _ := os.ReadFile(in)
	if _, err := ExtractXML(plain); err == nil {
		t.Error("plain pdf must error")
	}
}
