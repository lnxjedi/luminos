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
	"fmt"
	//"github.com/howeyc/fsnotify"
	"bytes"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/lnxjedi/luminos/page"
	"github.com/lnxjedi/luminos/watcher"
	"github.com/lnxjedi/to"
	"github.com/lnxjedi/yaml"
	"github.com/russross/blackfriday"
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

// Map of file extensions for files to serve up directly
func init() {
	directFileTypes = map[string]struct{}{
		".jpg":  struct{}{},
		".txt":  struct{}{},
		".png":  struct{}{},
		".jpeg": struct{}{},
		".svg":  struct{}{},
	}
}

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
	// Templates (not fully functional yet)
	Templates map[string]*template.Template
	// Function map for templates.
	template.FuncMap
	// Standard request.
	*http.Request
	// Standard response writer.
	http.ResponseWriter
	// Function map
	funcMap template.FuncMap
	// File watcher
	//Watcher *fsnotify.Watcher
	Watcher *watcher.Watcher
	// Template root
	TemplateRoot string
}

// Expected extensions. Elements on the left have precedence.
var extensions = []string{
	".md",
	".html",
	".txt",
	".md.tpl",
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
		path := host.DocumentRoot + pathSeparator + directory
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New(`content directory was not found`)
}

// readFile attempts to read a file from disk and returns its contents.
func readFile(file string) (string, error) {
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
func guessFile(file string, descend bool) (string, os.FileInfo) {
	stat, err := os.Stat(file)

	file = strings.TrimRight(file, pathSeparator)

	if descend {
		if err == nil {
			if stat.IsDir() {
				f, s := guessFile(file+pathSeparator+"index", true)
				if s != nil {
					return f, s
				}
			}
			return file, stat
		}
		for _, extension := range extensions {
			f, s := guessFile(file+extension, false)
			if s != nil {
				return f, s
			}
		}
	}

	if err == nil {
		return file, stat
	}

	return "", nil
}

// readFile opens a file and reads its contents, if the file has the .md
// extension the contents are parsed and HTML is returned.
func (host *Host) readFile(file string) ([]byte, error) {
	var buf []byte
	var err error

	if buf, err = ioutil.ReadFile(file); err != nil {
		return nil, err
	}

	if strings.HasSuffix(file, ".tpl") {
		var out bytes.Buffer
		tpl, err := template.New("").Funcs(host.funcMap).Parse(string(buf))
		if err != nil {
			return nil, err
		}
		if err := tpl.Execute(&out, nil); err != nil {
			return nil, err
		}
		file = file[:len(file)-4]
		buf = out.Bytes()
	}

	if strings.HasSuffix(file, ".md") {
		buf = blackfriday.Run(buf, blackfriday.WithExtensions(blackfriday.CommonExtensions|blackfriday.NoEmptyLineBeforeBlock))
	}

	return buf, nil
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
	webroot := host.DocumentRoot + pathSeparator + webrootdir

	// Attempt to match a request with a file in webroot/.
	localFile = webroot + pathSeparator + reqpath

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
		directFile := docroot + pathSeparator + reqpath

		stat, err = os.Stat(directFile)

		if err == nil {
			if stat.IsDir() == false {
				ext := path.Ext(directFile)
				_, direct := directFileTypes[ext]
				if direct {
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
		testFile := docroot + pathSeparator + reqpath

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
			content, err := host.readFile(localFile)

			if err == nil {
				p.Content = template.HTML(content)
			}

			p.FileDir = strings.TrimRight(p.FileDir, pathSeparator) + pathSeparator
			p.BasePath = strings.TrimRight(p.BasePath, pathSeparator) + pathSeparator

			// werc-like header and footer.
			hfile, hstat := guessFile(p.FileDir+"_header", true)

			if hstat != nil {
				hcontent, herr := host.readFile(hfile)
				if herr == nil {
					p.ContentHeader = template.HTML(hcontent)
				}
			}

			if strings.Trim(host.Path, pathSeparator) == strings.Trim(req.URL.Path, pathSeparator) {
				p.IsHome = true
			}
			fmt.Printf("FP: %s, FD: %s, BP: %s, BD: %s\n", p.FilePath, p.FileDir, p.BasePath, p.BaseDir)

			// werc-like header and footer.
			ffile, fstat := guessFile(p.FileDir+"_footer", true)

			if fstat != nil {
				fcontent, ferr := host.readFile(ffile)
				if ferr == nil {
					p.ContentFooter = template.HTML(fcontent)
				}
			}

			p.CreateBreadCrumb()
			p.CreateMenu()
			p.CreateSideMenu()

			p.ProcessContent()

			// Applying template.
			if err = host.Templates["index.tpl"].Execute(w, p); err != nil {
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

func (host *Host) loadTemplate(file string) error {
	var err error

	// Reading template file.
	var text string
	if text, err = readFile(file); err != nil {
		return err
	}

	// Fixing template.
	text = fixDeprecatedSyntax(text)

	// Allocating name.
	name := path.Base(file)
	parsed := template.New(name).Funcs(host.funcMap)

	if _, err = parsed.Parse(text); err != nil {
		return err
	}

	host.Templates[name] = parsed

	if host.Watcher != nil {
		host.Watcher.RemoveWatch(file)
		host.Watcher.Watch(file)
	}

	return nil
}

// loadTemplates loads templates with .tpl extension from the templates
// directory. At this moment only index.tpl is expected.
func (host *Host) loadTemplates() error {
	var err error
	var fp *os.File

	tpldir := to.String(host.Settings.Get("content", "templates"))

	if tpldir == "" {
		tpldir = "templates"
	}

	tplroot := host.DocumentRoot + pathSeparator + tpldir

	if fp, err = os.Open(tplroot); err != nil {
		return fmt.Errorf("Error trying to open %s: %q", tplroot, err)
	}

	defer fp.Close()

	host.TemplateRoot = tplroot

	var files []os.FileInfo
	if files, err = fp.Readdir(-1); err != nil {
		return fmt.Errorf("Error reading directory %s: %q", tplroot, err)
	}

	for _, fp := range files {

		if strings.HasSuffix(fp.Name(), ".tpl") == true {

			file := host.TemplateRoot + pathSeparator + fp.Name()

			err := host.loadTemplate(file)

			if err != nil {
				log.Printf("%s: Template error in file %s: %q\n", host.Name, file, err)
			}

		}
	}

	if _, ok := host.Templates["index.tpl"]; ok == false {
		return fmt.Errorf("Template %s could not be found.", "index.tpl")
	}

	return nil

}

func (host *Host) fileWatcher() error {

	var err error

	/*
		// File watcher.
		host.Watcher, err = fsnotify.NewWatcher()

		if err == nil {

			go func() {

				for {

					select {

					case ev := <-host.Watcher.Event:

						fmt.Printf("%s: got ev: %v\n", host.Name, ev)

						if ev == nil {
							return
						}

						if ev.IsModify() {
							// Is settings file?
							if ev.Name == host.DocumentRoot+pathSeparator+settingsFile {
								log.Printf("%s: Reloading host settings %s...\n", host.Name, ev.Name)
								err := host.loadSettings()

								if err != nil {
									log.Printf("%s: Could not reload host settings: %s\n", host.Name, host.DocumentRoot+pathSeparator+settingsFile)
								}
							}

							// Is a template?
							if strings.HasPrefix(ev.Name, host.TemplateRoot) == true {

								if strings.HasSuffix(ev.Name, ".tpl") == true {
									log.Printf("%s: Reloading template %s\n", host.Name, ev.Name)
									host.loadTemplate(ev.Name)

									if err != nil {
										log.Printf("%s: Could not reload template %s: %q\n", host.Name, ev.Name, err)
									}

								}
							}

						} else if ev.IsDelete() {
							// Attemping to re-add watcher.
							host.Watcher.RemoveWatch(ev.Name)
							host.Watcher.Watch(ev.Name)
						}

					}
				}

			}()

		}
	*/

	// (Stupid) file modification watcher.
	host.Watcher, err = watcher.New()

	if err == nil {

		go func() {

			for {
				select {
				case ev := <-host.Watcher.Event:

					if ev.IsModify() {
						// Is settings file?
						if ev.Name == host.DocumentRoot+pathSeparator+settingsFile {
							log.Printf("%s: Reloading host settings %s...\n", host.Name, ev.Name)
							err := host.loadSettings()

							if err != nil {
								log.Printf("%s: Could not reload host settings: %s\n", host.Name, host.DocumentRoot+pathSeparator+settingsFile)
							}
						}

						// Is a template?
						if strings.HasPrefix(ev.Name, host.TemplateRoot) == true {
							if strings.HasSuffix(ev.Name, ".tpl") == true {
								log.Printf("%s: Reloading template %s\n", host.Name, ev.Name)
								host.loadTemplate(ev.Name)
								if err != nil {
									log.Printf("%s: Could not reload template %s: %q\n", host.Name, ev.Name, err)
								}
							}
						}
					}
				}
			}
		}()
	}

	return err

}

// loadSettings loads settings for the host.
func (host *Host) loadSettings() error {

	var settings *yaml.Yaml

	file := host.DocumentRoot + pathSeparator + settingsFile

	_, err := os.Stat(file)

	if err == nil {
		settings, err = yaml.Open(file)
		if err != nil {
			return fmt.Errorf(`could not parse settings file (%s): %q`, file, err)
		}
	} else {
		return fmt.Errorf(`error trying to open settings file (%s): %q`, file, err)
	}

	if host.Watcher != nil {
		host.Watcher.RemoveWatch(file)
		host.Watcher.Watch(file)
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
		Templates:    make(map[string]*template.Template),
	}

	host.funcMap = template.FuncMap{
		"url":    func(s string) string { return host.url(s) },
		"anchor": func(a, b string) template.HTML { return host.anchor(a, b) },
		"asset":  func(s string) string { return host.asset(s) },
		"include": func(f string) string {
			s, err := readFile(host.DocumentRoot + "/" + f)
			if err != nil {
				log.Printf("readFile: %q", err)
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

	log.Printf("Routing: %s -> %s\n", name, root)

	return host, nil

}
