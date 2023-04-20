package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
)

var (
	dataDir  = "tmpdata"
	cacheDir = filepath.Join(dataDir, "cache")

	flgToken   string
	flgVerbose bool

	// if true, will try to avoid downloading the page by using
	// cached version saved in log/ directory
	flgNoCache bool

	// if true, will not automatically open a browser to display
	// html generated for a page
	flgNoOpen bool

	flgNoFormat bool
	flgReExport bool
)

func getToken() string {
	if flgToken != "" {
		return flgToken
	}
	return os.Getenv("NOTION_TOKEN")
}

func newClient() *notionapi.Client {
	c := &notionapi.Client{
		AuthToken: getToken(),
	}
	if flgVerbose {
		c.DebugLog = flgVerbose
		c.Logger = os.Stdout
	}
	return c
}

func exportPageToFile(id string, exportType string, recursive bool, path string) error {

	if exportType == "" {
		exportType = "html"
	}
	client := newClient()
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		logf("client.ExportPages() failed with '%s'\n", err)
		return err
	}

	writeFileMust(path, d)
	logf("Downloaded exported page of id %s as %s\n", id, path)
	return nil
}

func exportPage(id string, exportType string, recursive bool) {
	client := newClient()

	if exportType == "" {
		exportType = "html"
	}
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		logf("client.ExportPages() failed with '%s'\n", err)
		return
	}
	name := notionapi.ToNoDashID(id) + "-" + exportType + ".zip"
	writeFileMust(name, d)
	logf("Downloaded exported page of id %s as %s\n", id, name)
}

func runGoTests() {
	cmd := exec.Command("go", "test", "-v", "./...")
	logf("Running: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func traceNotionAPI() {
	nodeModulesDir := filepath.Join("tracenotion", "node_modules")
	if !u.DirExists(nodeModulesDir) {
		cmd := exec.Command("yarn")
		cmd.Dir = "tracenotion"
		err := cmd.Run()
		must(err)
	}
	scriptPath := filepath.Join("tracenotion", "trace.js")
	cmd := exec.Command("node", scriptPath)
	cmd.Args = append(cmd.Args, flag.Args()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	must(err)
}

var toText = notionapi.TextSpansToString

func main() {
	u.CdUpDir("notionapi")
	logf("currDirAbs: '%s'\n", u.CurrDirAbsMust())

	var (
		//flgToken string
		// id of notion page to download
		flgDownloadPage string

		// id of notion page to download and convert to HTML
		flgToHTML     string
		flgToMarkdown string

		flgPreviewHTML     string
		flgPreviewMarkdown string

		flgWc bool

		flgExportPage string
		flgExportType string
		flgRecursive  bool
		flgTrace      bool

		// if true, remove cache directories (data/log, data/cache)
		flgCleanCache bool

		flgSanityTest        bool
		flgSmokeTest         bool
		flgTestToMd          string
		flgTestToHTML        string
		flgTestDownloadCache string
		flgBench             bool
	)

	{
		flag.BoolVar(&flgNoFormat, "no-format", false, "if true, doesn't try to reformat/prettify HTML files during HTML testing")
		flag.BoolVar(&flgCleanCache, "clean-cache", false, "if true, cleans cache directories (data/log, data/cache")
		flag.StringVar(&flgToken, "token", "", "auth token")
		flag.BoolVar(&flgRecursive, "recursive", false, "if true, recursive export")
		flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
		flag.StringVar(&flgExportPage, "export-page", "", "id of the page to export")
		flag.BoolVar(&flgTrace, "trace", false, "run node tracenotion/trace.js")
		flag.StringVar(&flgExportType, "export-type", "", "html or markdown")
		flag.StringVar(&flgTestToMd, "test-to-md", "", "test markdown generation")
		flag.StringVar(&flgTestToHTML, "test-to-html", "", "id of start page")
		flag.StringVar(&flgToHTML, "to-html", "", "id of notion page to download and convert to html")
		flag.StringVar(&flgToMarkdown, "to-md", "", "id of notion page to download and convert to markdown")

		flag.StringVar(&flgPreviewHTML, "preview-html", "", "id of start page")
		flag.StringVar(&flgPreviewMarkdown, "preview-md", "", "id of start page")

		flag.BoolVar(&flgSanityTest, "sanity", false, "runs a quick sanity tests (fast and basic)")
		flag.BoolVar(&flgSmokeTest, "smoke", false, "run a smoke test (not fast, run after non-trivial changes)")
		flag.StringVar(&flgTestDownloadCache, "test-download-cache", "", "page id to use to test download cache")
		flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
		flag.BoolVar(&flgReExport, "re-export", false, "if true, will re-export from notion")
		flag.BoolVar(&flgNoCache, "no-cache", false, "if true, will not use a cached version in log/ directory")
		flag.BoolVar(&flgNoOpen, "no-open", false, "if true, will not automatically open the browser with html file generated with -tohtml")
		flag.BoolVar(&flgWc, "wc", false, "wc -l on source files")
		flag.BoolVar(&flgBench, "bench", false, "run benchmark")
		flag.Parse()
	}

	must(os.MkdirAll(cacheDir, 0755))

	if false {
		flgPreviewHTML = "da0b358c21ab4ac6b5c0f7154b2ecadc"
	}

	if false {
		testDownloadFile()
		return
	}

	if false {
		adhocTests()
		return
	}

	if false {
		testGetBlockRecords()
		testLoadCachePageChunk()
		return
	}

	if false {
		// simple page with an image
		pageID := "da0b358c21ab4ac6b5c0f7154b2ecadc"
		client := makeNotionClient()
		client.DebugLog = true
		if false {
			timeStart := time.Now()
			page, err := client.DownloadPage(pageID)
			if err != nil {
				logf("Client.DownloadPage('%s') failed with '%s'\n", pageID, err)
				return
			}
			logf("Client.DownloadPage('%s') downloaded page '%s' in %s\n", pageID, page.Root().GetTitle(), time.Since(timeStart))
		}
		// try with empty cache
		cacheDir, err := filepath.Abs("cached_notion")
		must(err)
		os.RemoveAll(cacheDir)
		logf("cache dir: '%s'\n", cacheDir)
		{
			client, err := notionapi.NewCachingClient(cacheDir, client)
			must(err)
			timeStart := time.Now()
			page, err := client.DownloadPage(pageID)
			if err != nil {
				logf("Client.DownloadPage('%s') failed with '%s'\n", pageID, err)
				return
			}
			logf("CachingClient.DownloadPage('%s') downloaded page '%s' in %s\n", pageID, toText(page.Root().GetTitle()), time.Since(timeStart))
			logf("Cached requests: %d, non-cached requests: %d, requests written to cache: %d\n", client.RequestsFromCache, client.RequestsFromServer, client.RequestsWrittenToCache)
		}
		// try with full cache
		{
			client, err := notionapi.NewCachingClient(cacheDir, client)
			must(err)
			timeStart := time.Now()
			page, err := client.DownloadPage(pageID)
			if err != nil {
				logf("Client.DownloadPage('%s') failed with '%s'\n", pageID, err)
				return
			}
			logf("CachingClient.DownloadPage('%s') downloaded page '%s' in %s\n", pageID, toText(page.Root().GetTitle()), time.Since(timeStart))
			logf("Cached requests: %d, non-cached requests: %d, requests written to cache: %d\n", client.RequestsFromCache, client.RequestsFromServer, client.RequestsWrittenToCache)
		}
		return
	}

	if false {
		// simple page with an image
		//flgToHTML = "da0b358c21ab4ac6b5c0f7154b2ecadc"
		//flgToHTML = "35fbba015f344570af678d56827dd67c"
		flgToHTML = "638829dcc8f24475afcdfa245d411e50"
	}

	if false {
		testSubPages()
		return
	}

	// normalize ids early on
	flgDownloadPage = notionapi.ToNoDashID(flgDownloadPage)
	flgToHTML = notionapi.ToNoDashID(flgToHTML)
	flgToMarkdown = notionapi.ToNoDashID(flgToMarkdown)

	if flgWc {
		doLineCount()
		return
	}

	if flgCleanCache {
		{
			dir := filepath.Join(dataDir, "diff")
			os.RemoveAll(dir)
		}
		{
			dir := filepath.Join(dataDir, "smoke")
			os.RemoveAll(dir)
		}
		u.RemoveFilesInDirMust(cacheDir)
	}

	if flgBench {
		cmd := exec.Command("go", "test", "-bench=.")
		u.RunCmdMust(cmd)
		return
	}

	if flgSanityTest {
		sanityTests()
		return
	}

	if flgSmokeTest {
		// smoke test includes sanity test
		sanityTests()
		smokeTest()
		return
	}

	if flgTrace {
		traceNotionAPI()
		return
	}

	if flgTestToMd != "" {
		testToMarkdown(flgTestToMd)
		return
	}

	if flgExportPage != "" {
		exportPage(flgExportPage, flgExportType, flgRecursive)
		return
	}

	if flgTestDownloadCache != "" {
		testCachingDownloads(flgTestDownloadCache)
		return
	}

	if flgTestToHTML != "" {
		testToHTML(flgTestToHTML)
		return
	}

	if flgDownloadPage != "" {
		client := makeNotionClient()
		downloadPage(client, flgDownloadPage)
		return
	}

	if flgToHTML != "" {
		flgNoCache = true
		toHTML(flgToHTML)
		return
	}

	if flgToMarkdown != "" {
		flgNoCache = true
		toMd(flgToMarkdown)
		return
	}

	if flgPreviewHTML != "" {
		uri := "/previewhtml/" + flgPreviewHTML
		startHTTPServer(uri)
		return
	}

	if flgPreviewMarkdown != "" {
		uri := "/previewmd/" + flgPreviewMarkdown
		startHTTPServer(uri)
		return
	}

	flag.Usage()
}

func startHTTPServer(uri string) {
	flgHTTPAddr := "localhost:8503"
	httpSrv := makeHTTPServer()
	httpSrv.Addr = flgHTTPAddr

	logf("Starting on addr: %v\n", flgHTTPAddr)

	chServerClosed := make(chan bool, 1)
	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		must(err)
		logf("HTTP server shutdown gracefully\n")
		chServerClosed <- true
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)

	openBrowser("http://" + flgHTTPAddr + uri)
	time.Sleep(time.Second * 2)

	sig := <-c
	logf("Got signal %s\n", sig)

	if httpSrv != nil {
		// Shutdown() needs a non-nil context
		_ = httpSrv.Shutdown(context.Background())
		select {
		case <-chServerClosed:
			// do nothing
		case <-time.After(time.Second * 5):
			// timeout
		}
	}

}
