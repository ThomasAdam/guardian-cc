package charts

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	htmltemplate "html/template"
	"os"
	"sort"
	"strconv"
	"strings"
	texttemplate "text/template"
	"time"
)

// ChartPlugin defines the interface each chart must implement.
type ChartPlugin interface {
	// Order returns the sort key for this chart section (e.g. "1", "2", "5a").
	Order() string
	// Render queries the DB and returns the rendered HTML fragment.
	Render(db *sql.DB, tmplDir string) (string, error)
}

// AllPlugins returns chart plugins in display order.
func AllPlugins() []ChartPlugin {
	return []ChartPlugin{
		&Chart1{},
		&Chart2{},
		&Chart3{},
		&Chart4{},
		&Chart5{},
		&Chart5a{},
		&Chart6{},
		&Chart7{},
		&Chart8{},
		&Chart9{},
		&Chart10{},
		&Chart11{},
		&Chart12{},
		&Chart13{},
		&Chart14{},
	}
}

// RenderAll runs all chart plugins and produces the final HTML page.
func RenderAll(db *sql.DB, tmplDir, outputFile string) error {
	plugins := AllPlugins()
	sections := make(map[string]htmltemplate.HTML)

	for _, p := range plugins {
		fmt.Fprintf(os.Stderr, "Looking at: chart%s...\n", p.Order())
		html, err := p.Render(db, tmplDir)
		if err != nil {
			return fmt.Errorf("rendering chart%s: %w", p.Order(), err)
		}
		sections[p.Order()] = htmltemplate.HTML(html)
	}

	// Sort section keys numerically so "2" < "5a" < "10" < "11".
	// strconv.Atoi("5a") returns 0, so we must parse only the leading digit run.
	leadingInt := func(s string) int {
		end := strings.IndexFunc(s, func(r rune) bool { return r < '0' || r > '9' })
		if end < 0 {
			end = len(s)
		}
		if end == 0 {
			return 0
		}
		n, _ := strconv.Atoi(s[:end])
		return n
	}
	keys := make([]string, 0, len(sections))
	for k := range sections {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni := leadingInt(keys[i])
		nj := leadingInt(keys[j])
		if ni != nj {
			return ni < nj
		}
		return keys[i] < keys[j]
	})

	mainData := struct {
		Sections  []htmltemplate.HTML
		Timestamp string
	}{
		Timestamp: time.Now().Format("Mon Jan 2 15:04:05 2006"),
	}
	for _, k := range keys {
		mainData.Sections = append(mainData.Sections, sections[k])
	}

	mainTmpl, err := htmltemplate.ParseFiles(tmplDir + "/main.tmpl")
	if err != nil {
		return fmt.Errorf("parsing main template: %w", err)
	}

	var buf bytes.Buffer
	if err := mainTmpl.Execute(&buf, mainData); err != nil {
		return fmt.Errorf("executing main template: %w", err)
	}

	if err := os.WriteFile(outputFile, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// --- Helpers ---

// barColorPalette is a 12-colour palette for single-series bar charts.
// Because c3's color.pattern applies per *series* (not per bar), we use
// data.color — a JS callback — to cycle colours across individual bars.
const barColorPalette = `["#4e79a7","#f28e2b","#e15759","#76b7b2","#59a14f","#edc948","#b07aa1","#ff9da7","#9c755f","#bab0ac","#d37295","#499894"]`

// barColorJS returns a JS snippet that patches a c3 chart instance (named
// jsVar) so each bar cycles through the palette by x-index.
// It must be injected after c3.generate() is called.
func barColorJS(jsVar string) string {
	return fmt.Sprintf(`(function(){
				var _pal = %s;
				%s.internal.color = function(d) { return _pal[d.index %% _pal.length]; };
				%s.flush();
			})();`, barColorPalette, jsVar, jsVar)
}

// toJSON converts a value to pretty-printed JSON for embedding in templates.
func toJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// writeAjaxFile writes a DataTables-compatible AJAX JSON file.
// It uses SetEscapeHTML(false) so that HTML tags (e.g. <a href=...>)
// embedded in the data are written literally, not escaped to \u003c etc.
func writeAjaxFile(filename string, data [][]string) error {
	wrapper := struct {
		Data [][]string `json:"data"`
	}{
		Data: data,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "   ")
	if err := enc.Encode(wrapper); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// toStringSlice converts a []any (as returned by DuckDB's LIST() aggregate
// via the go-duckdb driver) into a []string.
func toStringSlice(v any) []string {
	switch sl := v.(type) {
	case []any:
		out := make([]string, len(sl))
		for i, elem := range sl {
			out[i] = fmt.Sprintf("%v", elem)
		}
		return out
	case []string:
		return sl
	default:
		return nil
	}
}

// executeTemplate parses and executes a text/template file with data.
// We use text/template (not html/template) because chart sub-templates
// produce HTML fragments containing raw JavaScript and JSON that must
// not be escaped. The fragments are injected into the main page via
// html/template's template.HTML type, which prevents double-escaping.
func executeTemplate(tmplDir, tmplFile string, data any) (string, error) {
	funcMap := texttemplate.FuncMap{
		"toJSON": toJSON,
	}
	t, err := texttemplate.New(tmplFile).Funcs(funcMap).ParseFiles(tmplDir + "/" + tmplFile)
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", tmplFile, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %s: %w", tmplFile, err)
	}
	return buf.String(), nil
}
