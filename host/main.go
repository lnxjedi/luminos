// Copyright (c) 2012-2014 José Carlos Nieto, https://menteslibres.net/xiam
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package host

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/lnxjedi/luminos/page"
	"github.com/lnxjedi/to"
	"github.com/lnxjedi/yaml"
	"github.com/russross/blackfriday"

	jyaml "github.com/ghodss/yaml"
)

const (
	pathSeparator = string(os.PathSeparator)
	settingsFile  = "site.yaml"
)

var (
	// Used to guess when dealing with an external URL.
	isExternalLinkPattern = regexp.MustCompile(`^[a-zA-Z0-9]+:\/\/`)
	// Extensions to directly serve
	directFileTypes map[string]struct{}
)

// Host is the struct that represents virtual hosts.
type Host struct {
	// Host name
	Name string
	// Main directory
	DocumentRoot string
	// Main path
	Path string
	// Settings
	Settings *yaml.Yaml
	// Template
	TemplateGroup *template.Template
	// Lock for template
	*sync.RWMutex
	// Function map for template.
	template.FuncMap
	// Standard request.
	*http.Request
	// Standard response writer.
	http.ResponseWriter
	// Function map
	funcMap template.FuncMap
	// File watcher
	Watcher *fsnotify.Watcher
	// Template root
	TemplateRoot string
}

// Page frontmatter
type frontMatter struct {
	Template string
	// True when markdown content shouldn't be rendered
	Raw bool
}

type structuredContent struct {
	// Information extracted from page frontmatter
	pageInfo frontMatter
	// Page content without frontmatter
	Content []byte
}

// Expected extensions. Elements on the left have precedence.
var extensions = []string{
	".md",
	".html",
	".txt",
	".md.tpl",
}

// Informational and error logging to stderr; access log goes to stdout
func init() {
	log.SetOutput(os.Stderr)
}

// fixDeprecatedSyntax fixes old template syntax.
func fixDeprecatedSyntax(s string) string {

	s = strings.Replace(s, "{{ link", "{{ anchor", -1)
	s = strings.Replace(s, "{{link", "{{anchor", -1)
	s = strings.Replace(s, ".link", ".URL", -1)
	s = strings.Replace(s, ".url", ".URL", -1)
	s = strings.Replace(s, ".text", ".Text", -1)
	s = strings.Replace(s, "jstext", "js", -1)
	s = strings.Replace(s, "htmltext", "html", -1)

	return s
}

// GetContentPath provides the relative path to site content
func (host *Host) GetContentPath() (string, error) {
	var directories []string

	contentdir := to.String(host.Settings.Get("content", "markdown"))
	if contentdir == "" {
		directories = []string{
			"content",
			"markdown",
		}
	} else {
		directories = []string{contentdir}
	}

	for _, directory := range directories {
		path := path.Join(host.DocumentRoot, directory)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New(`content directory was not found`)
}

// readRawFile attempts to read a file from disk and returns its contents with no processing.
func readRawFile(file string) (string, error) {
	var buf []byte
	var err error

	if buf, err = ioutil.ReadFile(file); err != nil {
		return "", fmt.Errorf("Could not read file %s: %s", file, err.Error())
	}

	return string(buf), nil
}

// Close removes the watcher that is currently associated with the host.
func (host *Host) Close() {
	host.Watcher.Close()
}

// asset returns a relative URL.
func (host *Host) asset(assetURL string) string {
	if !host.isExternalLink(assetURL) {
		assetURL = strings.TrimLeft(assetURL, "/")
		p := strings.Trim(host.Path, "/")
		if p == "" {
			return "/" + assetURL
		}
		return "/" + p + "/" + assetURL
	}
	return assetURL
}

// url returns an absolute URL.
func (host *Host) url(url string) string {
	if host.isExternalLink(url) == false {
		return "//" + host.Request.Host + "/" + strings.TrimLeft(url, "/")
	}
	return url
}

// isExternalLink returns true if the given URL is outside this host.
func (host *Host) isExternalLink(url string) bool {
	return isExternalLinkPattern.MatchString(url)
}

// setting function returns a setting value.
func (host *Host) setting(path string) interface{} {
	route := strings.Split(path, "/")
	args := make([]interface{}, len(route))
	for i := range route {
		args[i] = route[i]
	}
	setting := host.Settings.Get(args...)
	return fixSetting(setting)
}

// fixSetting returns additional keys to make certain maps act like anchors.
func fixSetting(setting interface{}) interface{} {

	if m, ok := setting.(map[interface{}]interface{}); ok {
		for k, v := range m {
			switch k {
			case "text":
				m["Text"] = v
			case "link", "url":
				m["URL"] = v
			}
		}
	}

	return setting
}

// settings is a function that returns an array of settings.
func (host *Host) settings(path string) []interface{} {
	route := strings.Split(path, "/")
	args := make([]interface{}, len(route))
	for i := range route {
		args[i] = route[i]
	}
	val := host.Settings.Get(args...)
	if val == nil {
		return nil
	}

	ival := val.([]interface{})

	for i := range ival {
		ival[i] = fixSetting(ival[i])
	}

	return ival
}

// javascriptText is a function for funcMap that writes text as Javascript.
func javascriptText(text string) template.JS {
	return template.JS(text)
}

// htmlText is a function for funcMap that writes text as plain HTML.
func htmlText(text string) template.HTML {
	return template.HTML(text)
}

// Function for funcMap that writes links.
func (host *Host) anchor(url, text string) template.HTML {
	if host.isExternalLink(url) {
		return template.HTML(fmt.Sprintf(`<a target="_blank" href="%s">%s</a>`, host.asset(url), text))
	}
	return template.HTML(fmt.Sprintf(`<a href="%s">%s</a>`, host.asset(url), text))
}

// guessFile checks for files names and returns a guessed name.
func guessFile(file string, descend bool) (f string, s os.FileInfo) {
	var err error
	f = file
	// log.Printf("Guessing for '%s'", f)
	// defer func() {
	// 	log.Printf("Returning guessed file '%s'", f)
	// }()

	s, err = os.Stat(f)

	f = strings.TrimRight(file, pathSeparator)

	if descend {
		if err == nil {
			if s.IsDir() {
				idx, statidx := guessFile(path.Join(file, "index"), true)
				if statidx != nil {
					return idx, statidx
				}
			}
			return
		}
		for _, extension := range extensions {
			f, s = guessFile(file+extension, false)
			if s != nil {
				return
			}
		}
	}

	if err == nil {
		return
	}

	f = ""
	s = nil
	return
}

// readContentFile opens a file and reads its contents and frontmatter.
// If the file has the "*.md" extension, the content is rendered to HTML
// unless Raw is set in the frontmatter.
func (host *Host) readContentFile(file string) (structuredContent, error) {
	var sc structuredContent
	var buf []byte
	var err error

	if buf, err = ioutil.ReadFile(file); err != nil {
		return sc, err
	}

	if strings.HasSuffix(file, ".tpl") {
		var out bytes.Buffer
		tpl, err := template.New("").Funcs(host.funcMap).Parse(string(buf))
		if err != nil {
			return sc, err
		}
		if err := tpl.Execute(&out, nil); err != nil {
			return sc, err
		}
		file = file[:len(file)-4]
		buf = out.Bytes()
	} else {
		// maximum length of frontmatter start delimiter always < 16
		peek := make([]byte, 32)
		copy(peek, buf)
		r := bytes.NewBuffer(peek)
		firstLine, err := r.ReadString('\n')
		if err == nil {
			var end string
			switch firstLine {
			case "---\n":
				end = "---\n"
			case "<!--\n":
				end = "-->\n"
			case "```yaml\n":
				end = "```\n"
			}
			if len(end) != 0 {
				secondLine, _ := r.ReadString('\n')
				if secondLine == "#luminos\n" {
					r = bytes.NewBuffer(buf)
					r.ReadString('\n')
					var fmb bytes.Buffer
					for {
						line, err := r.ReadBytes('\n')
						if err != nil {
							return sc, fmt.Errorf("Unterminated frontmatter reading %s", file)
						}
						if string(line) == end {
							break
						}
						fmb.Write(line)
					}
					err := jyaml.Unmarshal(fmb.Bytes(), &sc.pageInfo)
					if err != nil {
						return sc, fmt.Errorf("Invalid frontmatter reading %s", file)
					}
					buf = r.Bytes()
				}
			}
		}
	}

	if strings.HasSuffix(file, ".md") && !sc.pageInfo.Raw {
		buf = blackfriday.Run(buf, blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.NoEmptyLineBeforeBlock))
	}

	sc.Content = buf

	return sc, nil
}

func chunk(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

// ServeHTTP reads a request and creates an appropriate response.
func (host *Host) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var localFile string

	// TODO: Fix this non-critical race condition.  We need to save some
	// variables in a per request basis, in particular the hostname. It may not
	// always match the host name we gave to it (i.e: the "default" hostname). A
	// per-request context would be useful.
	host.Request = req

	// Settings default status as not found.
	status := http.StatusNotFound

	// Default size is no size (-1).
	size := -1

	// Requested path
	reqpath := strings.TrimRight(req.URL.Path, "/")

	// Stripping path
	index := len(host.Path)

	// If the hosts contains a path and the request begins with the same path, it
	// is ignored for the matches.
	if reqpath[0:index] == host.Path {
		reqpath = reqpath[index:]
	}

	reqpath = strings.TrimRight(reqpath, "/")

	// Trying to match a file on webroot/
	webrootdir := to.String(host.Settings.Get("content", "webroot"))

	if webrootdir == "" {
		webrootdir = "webroot"
	}

	// Absolute local webroot.
	webroot := path.Join(host.DocumentRoot, webrootdir)

	// Attempt to match a request with a file in webroot/.
	localFile = path.Join(webroot, reqpath)

	stat, err := os.Stat(localFile)

	if err == nil {
		// File exists
		if stat.IsDir() == false {
			// Exists and it's not a directory, let's serve it.
			status = http.StatusOK // Changing status.
			http.ServeFile(w, req, localFile)
			size = int(stat.Size())
		}
	}

	if reqpath == "/search" {
		host.doSearch(w, req)
		return
	}

	// Absolute document root.
	var docroot string
	if docroot, err = host.GetContentPath(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if status == http.StatusNotFound {
		directFile := path.Join(docroot, reqpath)

		stat, err = os.Stat(directFile)

		if err == nil {
			if stat.IsDir() == false {
				ext := path.Ext(directFile)
				if ext != ".md" {
					status = http.StatusOK // Changing status.
					http.ServeFile(w, req, directFile)
					size = int(stat.Size())
				}
			}
		}
	}

	// Was the status already changed?
	if status == http.StatusNotFound {

		// Defining a filename to look for.
		testFile := path.Join(docroot, reqpath)

		localFile, stat = guessFile(testFile, true)

		if stat != nil {

			if reqpath != "" {
				// Let's not accept paths ending in "/".
				if stat.IsDir() == false {
					if strings.HasSuffix(req.URL.Path, "/") == true {
						http.Redirect(w, req, "/"+host.Path+"/"+reqpath, 301)
						w.Write([]byte(http.StatusText(301)))
						return
					}
				} else {
					if strings.HasSuffix(req.URL.Path, "/") == false {
						http.Redirect(w, req, req.URL.Path+"/", 301)
						w.Write([]byte(http.StatusText(301)))
						return
					}
				}
			}

			// Creating a page.
			p := &page.Page{}

			p.FilePath = localFile
			p.BasePath = req.URL.Path

			relPath := localFile[len(docroot):]

			if stat.IsDir() == false {
				p.FileDir = path.Dir(localFile)
				p.BasePath = path.Dir(relPath)
			} else {
				p.FileDir = localFile
				p.BasePath = relPath
			}

			// Reading contents.
			content, err := host.readContentFile(localFile)

			if err == nil {
				p.Content = template.HTML(content.Content)
			}

			p.FileDir = strings.TrimRight(p.FileDir, pathSeparator) + pathSeparator
			p.BasePath = strings.TrimRight(p.BasePath, pathSeparator) + pathSeparator

			// werc-like header and footer.
			hfile, hstat := guessFile(p.FileDir+"_header", true)

			if hstat != nil {
				hcontent, herr := host.readContentFile(hfile)
				if herr == nil {
					p.ContentHeader = template.HTML(hcontent.Content)
				}
			}

			if strings.Trim(host.Path, pathSeparator) == strings.Trim(req.URL.Path, pathSeparator) {
				p.IsHome = true
			}

			// werc-like header and footer.
			ffile, fstat := guessFile(p.FileDir+"_footer", true)

			if fstat != nil {
				fcontent, ferr := host.readContentFile(ffile)
				if ferr == nil {
					p.ContentFooter = template.HTML(fcontent.Content)
				}
			}

			p.CreateBreadCrumb()
			p.CreateMenu()
			p.CreateSideMenu()

			p.ProcessContent()

			host.RLock()
			ht := host.TemplateGroup
			host.RUnlock()
			// Applying template.
			template := "index.tpl"
			if len(content.pageInfo.Template) != 0 {
				if t := ht.Lookup(content.pageInfo.Template); t != nil {
					template = content.pageInfo.Template
				}
			}
			if err = ht.ExecuteTemplate(w, template, p); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				status = http.StatusInternalServerError
			} else {
				status = http.StatusOK
			}

		}
	}

	if status == http.StatusNotFound {
		http.Error(w, "Not found", http.StatusNotFound)
		fmt.Printf("Path not found: %s\n", reqpath)
	}

	// Log line.
	logLine := []string{
		chunk(req.RemoteAddr),
		chunk(""),
		chunk(""),
		chunk("[" + time.Now().Format("02/Jan/2006:15:04:05 -0700") + "]"),
		chunk("\"" + fmt.Sprintf("%s %s %s", req.Method, req.RequestURI, req.Proto) + "\""),
		chunk(fmt.Sprintf("%d", status)),
		chunk(fmt.Sprintf("%d", size)),
	}

	fmt.Println(strings.Join(logLine, " "))
}

// loadTemplates loads templates with .tpl extension from the templates
// directory. At this moment only index.tpl is expected.
func (host *Host) loadTemplates() error {

	tpldir := to.String(host.Settings.Get("content", "templates"))

	if tpldir == "" {
		tpldir = "templates"
	}

	tplroot := path.Join(host.DocumentRoot, tpldir)

	if _, err := os.Stat(tplroot); err != nil {
		return fmt.Errorf("Error checking template dir %s: %q", tplroot, err)
	}

	host.TemplateRoot = tplroot

	t := template.New(host.Name).Funcs(host.funcMap)
	wd, _ := os.Getwd()
	tglob := path.Join(wd, tplroot, "*.tpl")
	if _, err := t.ParseGlob(tglob); err != nil {
		log.Printf("Error parsing templates for %s: %v\n", host.Name, err)
	}
	for _, tpl := range t.Templates() {
		if tpl.Name() != host.Name {
			log.Printf("Parsed host %s template: %s\n", host.Name, tpl.Name())
		}
	}
	log.Printf("Loaded templates for %s from %s\n", host.Name, tplroot)
	host.Lock()
	host.TemplateGroup = t
	host.Unlock()

	if def := host.TemplateGroup.Lookup("index.tpl"); def == nil {
		return fmt.Errorf("default Template %s could not be found", "index.tpl")
	}

	return nil

}

func (host *Host) fileWatcher() error {

	var err error

	// File watcher.
	host.Watcher, err = fsnotify.NewWatcher()

	if err == nil {

		go func() {

			for {

				select {

				case ev, ok := <-host.Watcher.Events:

					if !ok {
						return
					}

					log.Printf("%s: got ev: %v\n", host.Name, ev)

					// Is settings file?
					if strings.HasSuffix(ev.Name, settingsFile) {
						log.Printf("%s: Reloading host settings %s...\n", host.Name, ev.Name)
						err := host.loadSettings()

						if err != nil {
							log.Printf("%s: Could not reload host settings: %s\n", host.Name, path.Join(host.DocumentRoot, settingsFile))
						}
					} else {
						if strings.HasSuffix(ev.Name, ".tpl") == true {
							log.Printf("%s: Reloading templates, %s changed", host.Name, ev.Name)
							host.loadTemplates()
						}
					}

				case err, ok := <-host.Watcher.Errors:
					if !ok {
						return
					}
					log.Println("fsnotify error:", err)
				}
			}

		}()
	}

	return err
}

// loadSettings loads settings for the host.
func (host *Host) loadSettings() error {

	var settings *yaml.Yaml

	file := path.Join(host.DocumentRoot, settingsFile)

	_, err := os.Stat(file)

	if err == nil {
		settings, err = yaml.Open(file)
		if err != nil {
			return fmt.Errorf(`could not parse settings file (%s): %q`, file, err)
		}
	} else {
		return fmt.Errorf(`error trying to open settings file (%s): %q`, file, err)
	}

	host.Settings = settings

	return nil
}

// New creates and returns a host.
func New(name string, root string) (*Host, error) {

	_, err := os.Stat(root)

	if err != nil {
		log.Printf("Error reading directory %s: %q\n", root, err)
		log.Printf("Checkout an example directory at https://github.com/xiam/luminos/tree/master/default\n")

		return nil, err
	}

	route := "/"
	name = strings.TrimRight(name, "/")

	index := strings.Index(name, "/")
	if index > -1 {
		route = name[index:]
	}

	host := &Host{
		Name:         strings.TrimRight(name, "/"),
		Path:         strings.TrimRight(route, "/"),
		DocumentRoot: root,
		RWMutex:      new(sync.RWMutex),
	}

	host.funcMap = template.FuncMap{
		"url":    func(s string) string { return host.url(s) },
		"anchor": func(a, b string) template.HTML { return host.anchor(a, b) },
		"asset":  func(s string) string { return host.asset(s) },
		"include": func(f string) string {
			s, err := readRawFile(path.Join(host.DocumentRoot, f))
			if err != nil {
				log.Printf("readRawFile: %q", err)
			}
			return s
		},
		"setting":  func(s string) interface{} { return host.setting(s) },
		"settings": func(s string) []interface{} { return host.settings(s) },
		"js":       javascriptText,
		"html":     htmlText,
	}

	// Watcher
	host.fileWatcher()
	// Watch settings file
	wd, _ := os.Getwd()
	sf := path.Join(wd, host.DocumentRoot, settingsFile)
	host.Watcher.Add(sf)

	// Loading host settings
	if err = host.loadSettings(); err != nil {
		log.Printf("Could not start host: %s\n", name)
		return nil, err
	}

	// Loading templates.
	if err = host.loadTemplates(); err != nil {
		log.Printf("Could not start host: %s\n", name)
		return nil, err
	}
	tpldir := to.String(host.Settings.Get("content", "templates"))
	if tpldir == "" {
		tpldir = "templates"
	}

	td := path.Join(wd, host.DocumentRoot, tpldir)
	host.Watcher.Add(td)

	log.Printf("Routing: %s -> %s\n", name, root)

	return host, nil

}
