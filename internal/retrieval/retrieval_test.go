package retrieval

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrivilegeFilteredBeforeRanking(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "public.md"), []byte("Reset from the clean recovery image."), 0600)
	os.WriteFile(filepath.Join(d, "priv.md"), []byte("Reset password using privileged override zebra zebra zebra."), 0600)
	m := Manifest{Documents: []Document{{Path: "public.md", Incident: "INC-1"}, {Path: "priv.md", Incident: "INC-1", Compartment: "legal"}}}
	got, err := Retrieve(d, m, Caller{ID: "sam", AssignedIncidents: []string{"INC-1"}}, "zebra password reset recovery", 5)
	if err != nil {
		t.Fatal(err)
	}
	if got.Refused {
		t.Fatalf("unexpected refusal: %+v", got)
	}
	for _, c := range got.Citations {
		if c.Source == "priv.md" {
			t.Fatal("privileged passage leaked")
		}
	}
}
func TestUnassignedRefused(t *testing.T) {
	d := t.TempDir()
	os.WriteFile(filepath.Join(d, "x.md"), []byte("insurer hotline details"), 0600)
	got, err := Retrieve(d, Manifest{Documents: []Document{{Path: "x.md", Incident: "INC-2"}}}, Caller{ID: "sam", AssignedIncidents: []string{"INC-1"}}, "insurer hotline", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Refused {
		t.Fatal("expected refusal")
	}
}
func TestDOCX(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "runbook.docx")
	f, _ := os.Create(p)
	zw := zip.NewWriter(f)
	x, _ := zw.Create("word/document.xml")
	x.Write([]byte(`<w:document xmlns:w="x"><w:body><w:p><w:r><w:t>Call</w:t><w:tab/><w:t>the insurer</w:t><w:br/><w:t>hotline first.</w:t></w:r></w:p></w:body></w:document>`))
	zw.Close()
	f.Close()
	got, err := Retrieve(d, Manifest{Documents: []Document{{Path: "runbook.docx"}}}, Caller{ID: "sam"}, "insurer hotline", 3)
	if err != nil {
		t.Fatal(err)
	}
	if got.Refused || got.Citations[0].Locator != "paragraph:1" || !strings.Contains(got.Citations[0].Passage, "Call the insurer hotline first") {
		t.Fatalf("bad result %+v", got)
	}
}
func TestPathEscapeRejected(t *testing.T) {
	_, err := Retrieve(t.TempDir(), Manifest{Documents: []Document{{Path: "../secret.md"}}}, Caller{ID: "a"}, "secret", 3)
	if err == nil {
		t.Fatal("expected path error")
	}
}
func TestSymlinkEscapeRejected(t *testing.T) {
	base := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.md"), []byte("secret material"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.md"), filepath.Join(base, "link.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	_, err := Retrieve(base, Manifest{Documents: []Document{{Path: "link.md"}}}, Caller{ID: "a"}, "secret", 3)
	if err == nil {
		t.Fatal("symlink escape accepted")
	}
}
func TestOversizedDOCXRejected(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "large.docx")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	x, err := zw.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := x.Write([]byte(`<w:document xmlns:w="x"><w:body><w:p><w:r><w:t>` + strings.Repeat("x", int(MaxDocumentBytes)+1) + `</w:t></w:r></w:p></w:body></w:document>`)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = Retrieve(d, Manifest{Documents: []Document{{Path: "large.docx"}}}, Caller{ID: "a"}, "xxxxx", 3)
	if err == nil {
		t.Fatal("oversized expanded DOCX accepted")
	}
}
func TestOversizedMarkdownRejected(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "large.md")
	if err := os.WriteFile(p, []byte(strings.Repeat("x", int(MaxDocumentBytes)+1)), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := Retrieve(d, Manifest{Documents: []Document{{Path: "large.md"}}}, Caller{ID: "a"}, "xxxxx", 3)
	if err == nil {
		t.Fatal("oversized Markdown accepted")
	}
}
func TestUnsupportedQueryRefuses(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "runbook.md"), []byte("Call the insurer hotline first."), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := Retrieve(d, Manifest{Documents: []Document{{Path: "runbook.md"}}}, Caller{ID: "sam"}, "satellite launch window", 3)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Refused || len(got.Citations) != 0 {
		t.Fatalf("unsupported query should refuse without citations: %+v", got)
	}
}
