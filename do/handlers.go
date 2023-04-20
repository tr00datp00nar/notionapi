package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi/tohtml"

	"github.com/kjk/notionapi/tomarkdown"
)

const (
	mimeTypeHTML       = "text/html; charset=utf-8"
	mimeTypeText       = "text/plain"
	mimeTypeJavaScript = "text/javascript; charset=utf-8"
	mimeTypeMarkdown   = "text/markdown; charset=UTF-8"
)

var (
	templates *template.Template
)

func reloadTemplates() {
	var err error
	pattern := filepath.Join("do", "*.tmpl.html")
	templates, err = template.ParseGlob(pattern)
	must(err)
}

func previewToMD(pageID string) ([]byte, error) {
	client := makeNotionClient()
	page, err := downloadPage(client, pageID)
	if err != nil {
		logf("previewToMD: downloadPage() failed with '%s'\n", err)
		return nil, err
	}
	if page == nil {
		logf("toHTML: page is nil\n")
		return nil, errors.New("page == nil")
	}
	conv := tomarkdown.NewConverter(page)
	// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
	// =>
	// /testmarkdown#${pageID}
	rewriteURL := func(uri string) string {
		logf("rewriteURL: '%s'", uri)
		// ExtractNoDashIDFromNotionURL() only checks if last part of the url
		// is a valid id. We only want to
		parsedURL, _ := url.Parse(uri)
		if !strings.Contains(uri, "notion.so") {
			logf("\n")
			return uri
		}
		//idStr := notionapi.ExtractNoDashIDFromNotionURL(uri)
		id := extractNotionIDFromURL(uri)
		if id == "" {
			if parsedURL != nil {
				//idStr = notionapi.ExtractNoDashIDFromNotionURL(parsedURL.Path)
				id = extractNotionIDFromURL(uri)
			}
			if id == "" {
				logf("\n")
				return uri
			}
		}

		res := "/previewmd/" + id
		logf("=> '%s'\n", res)
		// TODO: maybe preserve ?queryargs
		return res
	}

	conv.RewriteURL = rewriteURL
	d := conv.ToMarkdown()
	return d, nil
}

func previewToHTML(pageID string) ([]byte, error) {
	client := makeNotionClient()
	page, err := downloadPage(client, pageID)
	if err != nil {
		logf("previewToHTML: downloadPage() failed with '%s'\n", err)
		return nil, err
	}
	if page == nil {
		logf("toHTML: page is nil\n")
		return nil, errors.New("page == nil")
	}
	conv := tohtml.NewConverter(page)
	// change https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
	// =>
	// /previewhtml/${pageID}
	rewriteURL := func(uri string) string {
		logf("rewriteURL: '%s'", uri)
		// ExtractNoDashIDFromNotionURL() only checks if last part of the url
		// is a valid id. We only want to
		parsedURL, _ := url.Parse(uri)
		if !strings.Contains(uri, "notion.so") {
			logf("\n")
			return uri
		}
		//idStr := notionapi.ExtractNoDashIDFromNotionURL(uri)
		id := extractNotionIDFromURL(uri)
		if id == "" {
			if parsedURL != nil {
				//idStr = notionapi.ExtractNoDashIDFromNotionURL(parsedURL.Path)
				id = extractNotionIDFromURL(uri)
			}
			if id == "" {
				logf("\n")
				return uri
			}
		}

		res := "/previewhtml/" + id
		logf("=> '%s'\n", res)
		// TODO: maybe preserve ?queryargs
		return res
	}

	conv.RewriteURL = rewriteURL
	return conv.ToHTML()
}

func serveError(w http.ResponseWriter, r *http.Request, format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	w.Header().Set("Content-Type", mimeTypeText)
	code := http.StatusInternalServerError
	w.WriteHeader(code)
	_, _ = w.Write([]byte(s))
}

func serveHTMLTemplate(w http.ResponseWriter, r *http.Request, tmplName string, d interface{}) {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, tmplName, d)
	if err != nil {
		logf("tmpl.Execute failed with '%s'\n", err)
		return
	}
	w.Header().Set("Content-Type", mimeTypeHTML)
	code := http.StatusOK
	w.WriteHeader(code)
	_, _ = w.Write(buf.Bytes())
}

func handlePreviewHTML(w http.ResponseWriter, r *http.Request) {
	logf("handlePreviewHTML\n")
	reloadTemplates()

	pageID := extractNotionIDFromURL(r.URL.Path)
	if pageID == "" {
		logf("url '%s' has no valid notion id\n", r.URL)
		return
	}
	html, err := previewToHTML(pageID)
	if err != nil {
		logf("previewToHTML('%s') failed with '%s'\n", pageID, err)
		return
	}
	d := map[string]interface{}{
		"HTML": template.HTML(html),
	}
	serveHTMLTemplate(w, r, "preview.html.tmpl.html", d)
}

func handlePreviewMarkdown(w http.ResponseWriter, r *http.Request) {
	logf("handlePreviewMarkdown url: %s\n", r.URL)
	reloadTemplates()

	pageID := extractNotionIDFromURL(r.URL.Path)
	if pageID == "" {
		logf("url '%s' has no valid notion id\n", r.URL)
		return
	}
	md, err := previewToMD(pageID)
	if err != nil {
		logf("previewToMD('%s') failed with '%s'\n", pageID, err)
		return
	}

	// TODO: convert to HTML using some markdown library
	d := map[string]interface{}{
		"Markdown": string(md),
		"HTML":     template.HTML("<b>HTML preview</b>"),
	}
	serveHTMLTemplate(w, r, "preview.md.tmpl.html", d)
}

// https://blog.gopheracademy.com/advent-2016/exposing-go-on-the-internet/
func makeHTTPServer() *http.Server {
	mux := &http.ServeMux{}
	mux.HandleFunc("/previewhtml/", handlePreviewHTML)
	mux.HandleFunc("/previewmd/", handlePreviewMarkdown)
	var handler http.Handler = mux

	srv := &http.Server{
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second, // introduced in Go 1.8
		Handler:      handler,
	}
	return srv
}
