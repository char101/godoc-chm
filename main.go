package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/char101/godoc-chm/chm"
	path "github.com/char101/path.go"
	"golang.org/x/net/html"
)

type processFunc func(string, *goquery.Document)

var (
	styleRe             = regexp.MustCompile(`padding-left:\s*(\d+)px`)
	nbspPrefixRe        = regexp.MustCompile("^(\\s*(\u00A0|&nbsp;))*")
	nbspRe              = regexp.MustCompile("(\u00A0|&nbsp;)")
	absoluteURLRe       = regexp.MustCompile(`^(http|https|ftp)?://`)
	funcReceiverRe      = regexp.MustCompile(`^\(.+?\)`)
	project             = chm.NewProject("Go")
	cache               *Cache
	staticMap           = make(map[string]bool)
	blacklistedPrefixes = make([]string, 0)
	funcNameRe          = regexp.MustCompile(`^\w+`)
)

// fetch URL as string
func fetch(url string) []byte {
	if cache != nil {
		data := cache.get(url)
		if data != nil {
			return data
		}
	}
	log.Println("downloading", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if cache != nil {
		cache.set(url, body)
	}
	return body
}

func save(data interface{}, file string) {
	p := path.New(file)
	p.Dir().MkdirAll()

	switch v := data.(type) {
	case []byte:
		p.Write(v)
	case *goquery.Document:
		html, err := v.Html()
		if err != nil {
			log.Fatal(err)
		}
		p.Write(html)
	default:
		log.Fatalf("Unknown type: %T", v)
	}
}

func clean(url string, doc *goquery.Document) {
	fixPath := func(tag string, attr string) {
		doc.Find(tag).Each(func(i int, s *goquery.Selection) {
			val, _ := s.Attr(attr)
			if val != "" {
				if strings.HasPrefix(val, "/") && !strings.HasPrefix(val, "//") {
					s.SetAttr(attr, chm.RelativePath(url, val))
				} else {
					s.SetAttr(attr, chm.AddIndex(val))
				}
			}
		})
	}

	doc.Find("head").AppendHtml(`<link rel="stylesheet" href="/custom.css">`)

	fixPath("a", "href")
	fixPath("link[rel='stylesheet']", "href")
	fixPath("script", "src")
	fixPath("img", "src")
}

func parse(url string, process processFunc) (*goquery.Document, string) {
	var (
		file    = chm.GetFilename(url)
		content = fetch(url)
		reader  = strings.NewReader(string(content))
	)

	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		log.Fatal(err)
	}

	// process first then clean to keep the original URL
	if process != nil {
		process(url, doc)
	}
	downloadStatic(url, doc)

	clean(url, doc)

	save(doc, file)

	project.AddFile(file)

	return doc, file
}

func downloadStatic(baseURL string, doc *goquery.Document) {
	process := func(selector string, attr string) {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			url, _ := s.Attr(attr)
			if url != "" {
				url = chm.AbsoluteURL(baseURL, url)
				_, ok := staticMap[url]
				if !ok {
					file := chm.GetFilename(url)
					p := path.New(file)
					p.Dir().MkdirAll()
					p.Write(fetch(url))
				}
				staticMap[url] = true
			}
		})
	}
	process("link[rel='stylesheet']", "href")
	process("script", "src")
	process("img", "src")
}

func getTitle(doc *goquery.Document) string {
	return chm.CleanTitle(doc.Find("title").Text())
}

func isBlacklisted(pkg string) bool {
	for _, bl := range blacklistedPrefixes {
		if bl == pkg || strings.HasPrefix(pkg, bl+"/") {
			return true
		}
	}
	return false
}

// removes parameters and return values from function prototype
func simplifyFunc(f string) string {
	return fmt.Sprintf("%s()", funcNameRe.FindString(f))
}

func findIndex(toc *chm.TocItem, url string, doc *goquery.Document, pkg string) {
	var (
		prevLevel = 0
		currToc   = toc
		prevToc   *chm.TocItem
		index     = project.Index().Root()
		//index     = proj.Index().Root()
		getLevel = func(s *goquery.Selection) int {
			var (
				text    = s.Text()
				prefix  = nbspPrefixRe.FindString(text)
				matches = nbspRe.FindAllStringIndex(prefix, -1)
			)
			return len(matches) / 2
		}
	)
	log.Println(strings.Repeat("  ", toc.Level())+"findIndex:", url)

	h1 := doc.Find("#page h1")
	if strings.HasPrefix(strings.TrimSpace(h1.Text()), "Directory /") {
		return
	}

	doc.Find("#manual-nav dd").Each(func(i int, s *goquery.Selection) {
		level := getLevel(s)
		if level > prevLevel {
			currToc = prevToc
		} else if level < prevLevel {
			for i = level; i < prevLevel; i++ {
				currToc = currToc.Parent()
			}
		}

		a := s.Find("a")
		href, ok := a.Attr("href")
		if !ok {
			log.Fatal("href not found")
		}

		text := chm.CleanTitle(a.Text())
		link := strings.TrimPrefix(chm.AbsolutePath(url, href), "/")
		tag := ""
		if strings.HasPrefix(text, "type ") {
			tag = "type"
			text = text[5:]
			index.Add(fmt.Sprintf("%s (type in %s)", text, pkg)).AddLocal(link, pkg)
		} else if strings.HasPrefix(text, "func ") {
			text = text[5:]
			if strings.HasPrefix(text, "(") {
				tag = "method"
				text = strings.TrimSpace(funcReceiverRe.ReplaceAllString(text, ""))
				if !strings.HasPrefix(text, "String() string") {
					index.Add(fmt.Sprintf("%s (method of %s in %s)", simplifyFunc(text), currToc.Label(), pkg)).AddLocal(link, pkg)
				}
			} else {
				tag = "function"
				index.Add(fmt.Sprintf("%s (func in %s)", simplifyFunc(text), pkg)).AddLocal(link, pkg)
			}
		}

		t := currToc.Add(text, link)
		if tag != "" {
			t.TagAs(tag)
		}

		// add struct fields to the toc
		if tag == "type" {
			var id string
			var ft *chm.TocItem // fields toc, created as necessary
			doc.Find("h2#" + text).Next().Contents().Each(func(i int, s *goquery.Selection) {
				if id != "" && s.Get(0).Type == html.TextNode {
					if ft == nil {
						ft = t.Add("Fields", "")
					}
					tf := ft.Add(chm.CleanTitle(s.Text()), strings.TrimPrefix(chm.AbsolutePath(url, "#"+id), "/"))
					tf.TagAs("field")
					id = ""
				} else if goquery.NodeName(s) == "span" {
					id, _ = s.Attr("id")
				}
			})
		}

		if text == "Constants" {
			constants := doc.Find("#pkg-constants")
			curr := constants.Next()
			for curr.Length() > 0 && goquery.NodeName(curr) != "h2" {
				curr.Find("span").Each(func(i int, s *goquery.Selection) {
					if id, ok := s.Attr("id"); ok {
						text := chm.CleanTitle(s.Text())
						link := strings.TrimPrefix(chm.AbsolutePath(url, "#"+id), "/")
						t.Add(text, link)
						index.Add(fmt.Sprintf("%s(const in %s)", text, pkg)).AddLocal(link, pkg)
					}
				})
				curr = curr.Next()
			}
		} else if text == "Variables" {
			variables := doc.Find("#pkg-variables")
			curr := variables.Next()
			for curr.Length() > 0 && goquery.NodeName(curr) != "h2" {
				curr.Find("span").Each(func(i int, s *goquery.Selection) {
					if id, ok := s.Attr("id"); ok {
						text := chm.CleanTitle(s.Text())
						link := strings.TrimPrefix(chm.AbsolutePath(url, "#"+id), "/")
						t.Add(text, link)
						index.Add(fmt.Sprintf("%s(var in %s)", text, pkg)).AddLocal(link, pkg)
					}
				})
				curr = curr.Next()
			}
		}

		prevLevel = level
		prevToc = t
	})

	doc.Find("h3").Each(func(i int, h3 *goquery.Selection) {
		if h3.Text() == "Examples" {
			t := toc.Add("Examples", "")
			h3.Next().Find("a").Each(func(i int, a *goquery.Selection) {
				text := a.Text()
				href, _ := a.Attr("href")
				t.Add(text, strings.TrimPrefix(chm.AbsolutePath(url, href), "/"))
			})
		}
		if h3.Text() == "Package files" {
			t := toc.Add("Files", "")
			h3.Next().Find("a").Each(func(i int, a *goquery.Selection) {
				text := a.Text()
				href, _ := a.Attr("href")
				t.Add(text, strings.TrimPrefix(chm.AbsolutePath(url, href), "/"))

				// to download and clean the page
				parse(chm.AbsoluteURL(url, href), nil)
			})
		}
	})
}

func findPackages(url string, doc *goquery.Document) {
	var (
		prevLevel       = 0
		toc             = project.Toc().Root()
		prevToc         *chm.TocItem
		prevTitle       string
		prevBlacklisted bool
		index           = project.Index().Root()
		getLevel        = func(s *goquery.Selection) int {
			style, ok := s.Attr("style")
			if !ok {
				log.Fatal("style attribute not found")
			}
			matches := styleRe.FindStringSubmatch(style)
			if matches != nil {
				padding, err := strconv.Atoi(matches[1])
				if err != nil {
					log.Fatal(err)
				}
				return padding / 20
			}
			log.Fatal("cannot find padding")
			return 0
		}
		isDirectory = func(doc *goquery.Document) bool {
			h1 := doc.Find("#page h1")
			return strings.HasPrefix(strings.TrimSpace(h1.Text()), "Directory /")
		}
	)

	log.Println("findPackages", url)

	parents := make([]string, 0, 5)

	doc.Find("td.pkg-name").Each(func(i int, s *goquery.Selection) {
		level := getLevel(s)
		if level > prevLevel {
			if !prevBlacklisted {
				toc = prevToc
			}
			parents = append(parents, prevTitle)
		} else if level < prevLevel {
			for i = level; i < prevLevel; i++ {
				if !prevBlacklisted {
					toc = toc.Parent()
				}
				parents = parents[:len(parents)-1]
			}
		}

		a := s.Find("a")
		href, _ := a.Attr("href")

		title := chm.CleanTitle(a.Text())
		link := strings.TrimPrefix(chm.AbsolutePath(url, href), "/")

		fullPkg := strings.TrimPrefix(strings.Join(parents, "/")+"/"+title, "/")
		blacklisted := isBlacklisted(fullPkg)
		if blacklisted {
			log.Println(fullPkg, "is blacklisted")
		} else {
			tc := toc.Add(title, link)

			au := chm.AbsoluteURL(url, href)

			pkgdoc, _ := parse(au, func(url string, doc *goquery.Document) {
				findIndex(tc, url, doc, fullPkg)
			})

			if isDirectory(pkgdoc) {
				tc.TagAs("directory")
			} else {
				indexTitle := fmt.Sprintf("%s (package %s)", title, fullPkg)
				index.Add(indexTitle).AddLocal(link, getTitle(pkgdoc))
			}

			prevToc = tc
		}

		prevLevel = level
		prevTitle = title
		prevBlacklisted = blacklisted
	})
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var useCache bool
	flag.BoolVar(&useCache, "cache", false, "Cache request responses in a database")

	var outputDir string
	flag.StringVar(&outputDir, "output", "", "Output directory for downloaded files")

	var blacklist string
	flag.StringVar(&blacklist, "blacklist", "", "Blacklisted prefixes, separated by comma")

	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] godoc-url\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if blacklist != "" {
		for _, bl := range strings.Split(blacklist, "/") {
			blacklistedPrefixes = append(blacklistedPrefixes, strings.TrimSpace(bl))
		}
	}

	godocURL := flag.Arg(0)
	if strings.HasSuffix(godocURL, "/pkg") {
		godocURL += "/"
	} else if !strings.HasSuffix(godocURL, "/pkg/") {
		godocURL += "/pkg/"
	}

	if outputDir != "" {
		outputDir, err := filepath.Abs(outputDir)
		if err != nil {
			log.Fatal(err)
		}
		path.New(outputDir).MkdirAll().Chdir()
	}

	if useCache {
		cache = newCache()
		defer cache.close()
	}

	project.Toc().Root().Add("Packages", "pkg/index.html")
	project.SetStartFile("pkg/index.html")
	parse(godocURL, findPackages)
	if outputDir != "" {
		exe, err := os.Executable()
		if err != nil {
			log.Fatal(err)
		}
		chm.LinkFile(path.New(exe).Dir().Join("custom.css").String(), outputDir)
	}
	project.AddFile("custom.css")
	project.Save()
}
