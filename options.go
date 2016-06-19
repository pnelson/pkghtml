package pkghtml

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/pnelson/pkgdoc"
)

// Option describes a functional option for configuring the file system.
type Option func(*handler)

// Render sets the package documentation rendering function.
// Defaults to rendering a default template.
func Render(fn func(doc pkgdoc.Package) ([]byte, error)) Option {
	return func(h *handler) {
		h.render = fn
	}
}

// Template sets the package documentation template. The default renderer
// applies this template to an instance of pkgdoc.Package. This option is
// ignored if the Render option is used. Defaults to a simple HTML5 document.
func Template(filename string) Option {
	return func(h *handler) {
		b, err := ioutil.ReadFile(filename)
		if err == nil {
			h.template = string(b)
		}
	}
}

// StylesheetURL sets the stylesheet to apply to the package documentation.
// Templates rendered by the default renderer can link to this stylesheet.
// No stylesheet is applied to the default template if this value is empty.
// Defaults to the empty string.
func StylesheetURL(href string) Option {
	return func(h *handler) {
		h.stylesheet = href
	}
}

// ErrorHandler sets the http.Handler to delegate to when errors are returned.
// Defaults to writing a response with HTTP 404 Not Found if the package fails
// to import, otherwise HTTP 500 Internal Server Error to the response.
func ErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) Option {
	return func(h *handler) {
		h.errorHandler = fn
	}
}

// UpdateDuration sets the background update duration.
// Defaults to one hour.
func UpdateDuration(d time.Duration) Option {
	return func(h *handler) {
		h.updateDuration = d
	}
}

// defaultStylesheet is the default stylesheet for the default renderer.
const defaultStylesheet = ""

// defaultTemplate is the default template for the default renderer.
const defaultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>{{.Name}}</title>
<meta name="description" content="{{.Synopsis}}">
<meta name="viewport" content="width=device-width, initial-scale=1">
{{if .StylesheetURL}}<link rel="stylesheet" href="{{.StylesheetURL}}">{{end -}}
</head>
<body>
<h1>{{.Name}}</h1>
<p>{{.ImportPath}}</p>
{{.Doc.HTML}}

<h2 id="index">Index</h2>
<ul>
{{- if .Constants -}}
<li><a href="#constants">Constants</a></li>
{{- end -}}
{{- if .Variables -}}
<li><a href="#variables">Variables</a></li>
{{- end -}}
{{- range .Functions -}}
<li><a href="#{{.Name}}">{{.Decl}}</a></li>
{{- end -}}
{{- range $t := .Types -}}
<li><a href="#{{.Name}}">type {{.Name}}</a></li>
{{- if or .Functions .Methods -}}
<ul>
{{- end -}}
{{- range .Functions -}}
<li><a href="#{{.Name}}">{{.Decl}}</a></li>
{{- end -}}
{{- range .Methods -}}
<li><a href="#{{$t.Name}}.{{.Name}}">{{.Decl}}</a></li>
{{- end -}}
{{- if or .Functions .Methods -}}
</ul>
{{- end -}}
{{- end -}}
</ul>

{{- with .Constants -}}
<h2 id="constants">Constants</h2>
{{- range . -}}
<pre>{{.Decl}}</pre>
{{.Doc.HTML}}
{{- end -}}
{{- end -}}

{{- with .Variables -}}
<h2 id="variables">Variables</h2>
{{- range . -}}
<pre>{{.Decl}}</pre>
{{.Doc.HTML}}
{{- end -}}
{{- end -}}

{{- range .Functions -}}
<h2 id="{{.Name}}">{{.Decl}}</h2>
{{.Doc.HTML}}
{{- end -}}

{{- range $t := .Types -}}
<h2 id="{{.Name}}">type {{.Name}}</h2>
<pre>{{.Decl}}</pre>
{{.Doc.HTML}}
{{- range .Constants -}}
<pre>{{.Decl}}</pre>
{{.Doc.HTML}}
{{- end -}}
{{- range .Variables -}}
<pre>{{.Decl}}</pre>
{{.Doc.HTML}}
{{- end -}}
{{- range .Functions -}}
<h3 id="{{.Name}}">{{.Decl}}</h3>
{{.Doc.HTML}}
{{- end -}}
{{- range .Methods -}}
<h3 id="{{$t.Name}}.{{.Name}}">{{.Decl}}</h3>
{{.Doc.HTML}}
{{- end -}}
{{- end -}}

{{- with .SubPackages -}}
<h2 id="subpackages">Subpackages</h2>
<ul>
{{- range . -}}
<li><a href="{{.}}/">{{.}}</a></li>
{{- end -}}
</ul>
{{- end -}}
</body>
</html>`
