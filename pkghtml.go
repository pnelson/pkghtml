// Package pkghtml implements a http.Handler that renders package documentation.
package pkghtml

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/pnelson/pkgdoc"
)

// handler represents a http.Handler that renders package documentation.
type handler struct {
	mu             sync.Mutex
	name           string
	render         func(doc pkgdoc.Package) ([]byte, error)
	template       string
	stylesheet     string
	errorHandler   func(w http.ResponseWriter, req *http.Request, err error)
	updateDuration time.Duration
	packages       map[string]*pkg
}

// pkg represents package documentation.
type pkg struct {
	buf     []byte
	modTime time.Time
}

func (p pkg) getReadSeeker() io.ReadSeeker {
	return bytes.NewReader(p.buf)
}

// ErrImport represents an import error. This is useful
// for custom error handlers.
var ErrImport = errors.New("pkghtml: import failed")

// New returns a http.Handler that renders package documentation.
func New(name string, opts ...Option) http.Handler {
	h := &handler{
		name:           name,
		template:       defaultTemplate,
		stylesheet:     defaultStylesheet,
		errorHandler:   defaultErrorHandler,
		updateDuration: time.Hour,
		packages:       make(map[string]*pkg),
	}
	for _, option := range opts {
		option(h)
	}
	if h.render == nil {
		h.render = h.defaultRenderer
	}
	return h
}

// ServeHTTP implements the http.Handler interface.
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	url := req.URL.Path
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	if url[len(url)-1] != '/' {
		q := req.URL.RawQuery
		loc := path.Base(url) + "/"
		if q != "" {
			loc += "?" + q
		}
		w.Header().Set("Location", loc)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	name := h.name
	if url != "/" {
		name += path.Clean(url)
	}
	buf, modTime, err := h.prepare(name)
	if err != nil {
		h.errorHandler(w, req, err)
		return
	}
	http.ServeContent(w, req, name, modTime, buf)
}

func (h *handler) prepare(name string) (io.ReadSeeker, time.Time, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	p, ok := h.packages[name]
	if !ok {
		var err error
		p, err = h.fetch(name)
		if err != nil {
			return nil, time.Time{}, err
		}
		go h.update(name)
		h.packages[name] = p
	}
	return p.getReadSeeker(), p.modTime, nil
}

func (h *handler) fetch(name string) (*pkg, error) {
	doc, err := pkgdoc.New(name)
	if doc.Name == "" {
		err = ErrImport
	}
	if err != nil {
		return nil, err
	}
	b, err := h.render(doc)
	if err != nil {
		return nil, err
	}
	return &pkg{buf: b, modTime: time.Now()}, nil
}

func (h *handler) update(name string) {
	for {
		select {
		case <-time.After(h.updateDuration):
			p, err := h.fetch(name)
			if err != nil {
				continue
			}
			h.mu.Lock()
			if !bytes.Equal(h.packages[name].buf, p.buf) {
				h.packages[name] = p
			}
			h.mu.Unlock()
		}
	}
}

func (h *handler) defaultRenderer(doc pkgdoc.Package) ([]byte, error) {
	t, err := template.New("doc").Parse(h.template)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	type data struct {
		pkgdoc.Package
		StylesheetURL string
	}
	err = t.Execute(&buf, data{Package: doc, StylesheetURL: h.stylesheet})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func defaultErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
	var code int
	switch err {
	case ErrImport:
		code = http.StatusNotFound
	default:
		code = http.StatusInternalServerError
	}
	http.Error(w, http.StatusText(code), code)
}
