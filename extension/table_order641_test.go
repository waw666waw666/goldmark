package extension

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/parser"
	"github.com/yuin/goldmark/v2/renderer/html"
)

func renderGFM(t *testing.T, src string) (string, []string) {
	t.Helper()
	p := parser.New(parser.WithExtensions(NewGFMParser()))
	r := html.New(html.WithExtensions(NewGFMHTMLRenderer()))
	b := []byte(src)
	doc := p.Parse(b)
	var kinds []string
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		kinds = append(kinds, fmt.Sprintf("%s@%d", c.Kind().String(), c.Pos()))
	}
	var buf bytes.Buffer
	if err := r.Render(&buf, b, doc); err != nil {
		t.Fatal(err)
	}
	return buf.String(), kinds
}

// Issue #641: a delimiter row should take only the last paragraph line as the
// table header; lines above stay a paragraph before the table, and a following
// lone "-" becomes a list item — not a setext heading after the table.
func TestTableParagraphSplitSourceOrder(t *testing.T) {
	out, kinds := renderGFM(t, "0\n0\n-|\n-")
	t.Logf("children: %v", kinds)
	t.Logf("HTML: %q", out)
	if len(kinds) > 0 && strings.HasPrefix(kinds[0], "Heading") {
		t.Fatalf("first child is Heading (out of source order): %v", kinds)
	}
	for _, k := range kinds {
		if strings.HasPrefix(k, "Heading") {
			t.Fatalf("unexpected Heading among children %v; html=%q", kinds, out)
		}
	}
	if !strings.Contains(out, "<p>0</p>") {
		t.Errorf("expected leading paragraph; got %q", out)
	}
	if !strings.Contains(out, "<table>") || !strings.Contains(out, "<th>0</th>") {
		t.Errorf("expected table header 0; got %q", out)
	}
	if !strings.Contains(out, "<ul>") {
		t.Errorf("expected list for trailing '-'; got %q", out)
	}
}

// Same shape as the RequireParagraph comment in parser.go: header+delim then a
// lone '-' must become an empty list, not a setext heading.
func TestTableThenEmptyListNotSetext(t *testing.T) {
	out, kinds := renderGFM(t, "0\n-:\n-")
	t.Logf("children: %v", kinds)
	t.Logf("HTML: %q", out)
	for _, k := range kinds {
		if strings.HasPrefix(k, "Heading") {
			t.Fatalf("unexpected Heading among children %v; html=%q", kinds, out)
		}
	}
	if !strings.Contains(out, "<table>") {
		t.Errorf("expected table; got %q", out)
	}
	if !strings.Contains(out, "<ul>") {
		t.Errorf("expected list; got %q", out)
	}
}

// Multi-line residual above the header must stay one paragraph before the table.
func TestTableMultiLineResidualBeforeHeader(t *testing.T) {
	out, kinds := renderGFM(t, "a\nb\nc\n-|\n")
	t.Logf("children: %v", kinds)
	t.Logf("HTML: %q", out)
	if !strings.HasPrefix(kinds[0], "Paragraph") {
		t.Fatalf("expected leading paragraph, got %v", kinds)
	}
	if !strings.Contains(out, "<p>a\nb</p>") {
		t.Errorf("expected residual paragraph a/b; got %q", out)
	}
	if !strings.Contains(out, "<th>c</th>") {
		t.Errorf("expected header c; got %q", out)
	}
}
