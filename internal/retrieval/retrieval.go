package retrieval

import (
	"archive/zip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const MaxDocumentBytes int64 = 8 << 20

type Document struct {
	Path        string `json:"path"`
	Incident    string `json:"incident,omitempty"`
	Compartment string `json:"compartment,omitempty"`
}

type Caller struct {
	ID                string   `json:"id"`
	AssignedIncidents []string `json:"assigned_incidents,omitempty"`
	Compartments      []string `json:"compartments,omitempty"`
}

type Manifest struct {
	Documents []Document `json:"documents"`
}

type Citation struct {
	Source  string `json:"source"`
	Locator string `json:"locator"`
	Passage string `json:"passage"`
	Score   int    `json:"score"`
}

type Result struct {
	Refused   bool       `json:"refused"`
	Reason    string     `json:"reason,omitempty"`
	Citations []Citation `json:"citations,omitempty"`
}

type passage struct{ source, locator, text string }

func Retrieve(baseDir string, m Manifest, caller Caller, query string, limit int) (Result, error) {
	if strings.TrimSpace(query) == "" {
		return Result{}, errors.New("query is required")
	}
	if strings.TrimSpace(caller.ID) == "" {
		return Result{}, errors.New("caller id is required")
	}
	if limit <= 0 {
		limit = 3
	}
	q := tokens(query)
	if len(q) == 0 {
		return Result{Refused: true, Reason: "query contains no searchable terms"}, nil
	}
	var hits []Citation
	for _, d := range m.Documents {
		if !allowed(d, caller) {
			continue
		}
		ps, err := loadDocument(baseDir, d.Path)
		if err != nil {
			return Result{}, fmt.Errorf("%s: %w", d.Path, err)
		}
		for _, p := range ps {
			score := overlap(q, tokens(p.text))
			if score > 0 {
				hits = append(hits, Citation{Source: d.Path, Locator: p.locator, Passage: p.text, Score: score})
			}
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			if hits[i].Source == hits[j].Source {
				return hits[i].Locator < hits[j].Locator
			}
			return hits[i].Source < hits[j].Source
		}
		return hits[i].Score > hits[j].Score
	})
	if len(hits) == 0 {
		return Result{Refused: true, Reason: "no authorised passage supports the query"}, nil
	}
	if len(hits) > limit {
		hits = hits[:limit]
	}
	return Result{Citations: hits}, nil
}

func allowed(d Document, c Caller) bool {
	if d.Incident != "" && !contains(c.AssignedIncidents, d.Incident) {
		return false
	}
	if d.Compartment != "" && !contains(c.Compartments, d.Compartment) {
		return false
	}
	return true
}
func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func tokens(s string) map[string]struct{} {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, s)
	stop := map[string]bool{"the": true, "a": true, "an": true, "and": true, "or": true, "to": true, "of": true, "in": true, "on": true, "for": true, "is": true, "are": true, "what": true, "when": true, "how": true}
	out := map[string]struct{}{}
	for _, f := range strings.Fields(s) {
		if len(f) > 1 && !stop[f] {
			out[f] = struct{}{}
		}
	}
	return out
}
func overlap(a, b map[string]struct{}) int {
	n := 0
	for x := range a {
		if _, ok := b[x]; ok {
			n++
		}
	}
	return n
}

func loadDocument(base, path string) ([]passage, error) {
	full, display, err := containedDocumentPath(base, path)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(filepath.Ext(display)) {
	case ".md", ".markdown":
		return loadMarkdown(full, display)
	case ".docx":
		return loadDOCX(full, display)
	default:
		return nil, fmt.Errorf("unsupported format %q", filepath.Ext(display))
	}
}

func containedDocumentPath(base, name string) (string, string, error) {
	clean := filepath.Clean(name)
	if strings.TrimSpace(name) == "" || clean == "." || filepath.IsAbs(clean) {
		return "", "", errors.New("document path must stay under corpus directory")
	}
	baseAbs, err := filepath.Abs(base)
	if err != nil {
		return "", "", err
	}
	baseResolved, err := filepath.EvalSymlinks(baseAbs)
	if err != nil {
		return "", "", err
	}
	candidate := filepath.Join(baseAbs, clean)
	rel, err := filepath.Rel(baseAbs, candidate)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", "", errors.New("document path must stay under corpus directory")
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", "", err
	}
	relResolved, err := filepath.Rel(baseResolved, resolved)
	if err != nil || relResolved == ".." || strings.HasPrefix(relResolved, ".."+string(filepath.Separator)) || filepath.IsAbs(relResolved) {
		return "", "", errors.New("document path resolves outside corpus directory")
	}
	return resolved, filepath.ToSlash(rel), nil
}

func loadMarkdown(full, display string) ([]passage, error) {
	b, err := readBoundedFile(full, MaxDocumentBytes)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(b), "\n")
	var out []passage
	start := 0
	var buf []string
	flush := func(end int) {
		if len(buf) == 0 {
			return
		}
		text := strings.TrimSpace(strings.Join(buf, " "))
		if text != "" {
			out = append(out, passage{display, fmt.Sprintf("lines:%d-%d", start+1, end), text})
		}
		buf = nil
	}
	for i, l := range lines {
		if strings.TrimSpace(l) == "" {
			flush(i)
			continue
		}
		if len(buf) == 0 {
			start = i
		}
		buf = append(buf, strings.TrimSpace(l))
	}
	flush(len(lines))
	return out, nil
}

func readBoundedFile(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > limit {
		return nil, fmt.Errorf("document exceeds %d bytes", limit)
	}
	b, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("document exceeds %d bytes", limit)
	}
	return b, nil
}

func loadDOCX(full, display string) ([]passage, error) {
	st, err := os.Stat(full)
	if err != nil {
		return nil, err
	}
	if st.Size() > MaxDocumentBytes {
		return nil, fmt.Errorf("DOCX archive exceeds %d bytes", MaxDocumentBytes)
	}
	z, err := zip.OpenReader(full)
	if err != nil {
		return nil, err
	}
	defer z.Close()
	var f *zip.File
	for _, x := range z.File {
		if x.Name == "word/document.xml" {
			f = x
			break
		}
	}
	if f == nil {
		return nil, errors.New("DOCX missing word/document.xml")
	}
	if f.UncompressedSize64 > uint64(MaxDocumentBytes) {
		return nil, fmt.Errorf("DOCX document.xml exceeds %d bytes expanded", MaxDocumentBytes)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return parseDOCX(io.LimitReader(rc, MaxDocumentBytes+1), display)
}

func parseDOCX(r io.Reader, display string) ([]passage, error) {
	dec := xml.NewDecoder(r)
	var out []passage
	var para strings.Builder
	inText := false
	n := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case "tab", "br", "cr":
				if para.Len() > 0 {
					para.WriteByte(' ')
				}
			}
		case xml.CharData:
			if inText {
				if para.Len() > 0 && !strings.HasSuffix(para.String(), " ") {
					para.WriteByte(' ')
				}
				para.Write([]byte(t))
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
			if t.Name.Local == "p" {
				text := strings.Join(strings.Fields(para.String()), " ")
				para.Reset()
				if text != "" {
					n++
					out = append(out, passage{display, fmt.Sprintf("paragraph:%d", n), text})
				}
			}
		}
	}
	return out, nil
}
