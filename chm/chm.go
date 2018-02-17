package chm

import (
	"log"
	"net/url"
	"regexp"
	"strings"

	path "github.com/char101/path.go"
)

// AbsoluteURL creates an absolute URL from a base URL and a relative URL
func AbsoluteURL(base, href string) string {
	bu, err := url.Parse(base)
	if err != nil {
		log.Fatal(err)
	}

	uu, err := url.Parse(href)
	if err != nil {
		log.Fatal(err)
	}

	u := bu.ResolveReference(uu)

	return u.String()
}

func parseURL(u string) *url.URL {
	ur, err := url.Parse(u)
	if err != nil {
		log.Fatal(err)
	}
	return ur
}

func AddIndex(u interface{}) string {
	var p *url.URL
	switch v := u.(type) {
	case string:
		p = parseURL(v)
	case *url.URL:
		p = v
	default:
		log.Fatalf("Unsupported type: %T", v)
	}
	if strings.HasSuffix(p.Path, "/") {
		p.Path += "index.html"
	}
	return p.String()
}

// AbsolutePath converts a relative URL to an absolute one to be used in toc and index path
func AbsolutePath(base, href string) string {
	bu := parseURL(base)
	uu := parseURL(href)

	// external link
	if uu.Host != "" && bu.Host != uu.Host {
		return uu.String()
	}

	u := bu.ResolveReference(uu)

	// clear the host path
	u.Scheme = ""
	u.Host = ""

	return AddIndex(u)
}

// RelativePath rewrites URL inside a HTML file to use relative path
func RelativePath(base, resource string) string {
	if strings.HasPrefix(resource, "/") && !strings.HasPrefix(resource, "//") {
		basePath := parseURL(base).Path
		r := path.New(basePath).Dir().RelOf(resource)
		return AddIndex(strings.Replace(string(r), `\`, "/", -1))
	}
	return AddIndex(resource)
}

var (
	multipleSpaceRe = regexp.MustCompile(`\s{2,}`)
	newlineRe       = regexp.MustCompile(`\r|\n|\t`)
)

// CleanTitle returns a cleaned up text for toc & index titles
func CleanTitle(t string) string {
	t = newlineRe.ReplaceAllString(t, " ")
	t = multipleSpaceRe.ReplaceAllString(t, " ")
	t = strings.TrimSpace(t)
	return t
}

// GetFilename returns a relative local filename from URL
func GetFilename(u string) string {
	ur := parseURL(u)

	f := ur.Path
	if strings.HasSuffix(f, "/") {
		f += "index.html"
	}

	return strings.TrimPrefix(f, "/")
}

// CopyFile copies additional files (custom style, etc.) to the output directory
// but only if the destination file does not exist or the files are different
// based on modified time and size.
func CopyFile(src string, dst string) {
	sp := path.New(src)
	dp := path.New(dst)
	if !dp.Exists() || sp.Size() != dp.Size() || sp.ModTime() != dp.ModTime() {
		sp.Copy(dp)
	}
}

// LinkFile softlink a file to the destination directory
func LinkFile(src string, dst string) {
	sp := path.New(src)
	dp := path.New(dst)
	if dp.IsDir() {
		dp = dp.Join(sp.Basename())
	}
	if !dp.Exists() {
		_, err := sp.SymlinkErr(dp)
		if err != nil {
			log.Printf("Symlinking %s to %s %v failed (%v), copying instead", sp, dp, dp.IsDir(), err)
			sp.Copy(dp)
		}
	}
}
