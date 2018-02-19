package chm

import (
	"log"
	"regexp"
	"sort"
	"strings"
)

// children of a keyword is indented
// multiple topics of a keyword is displayed in a popup window

// Local stores data for a keyword local
type Local struct {
	href  string
	title string
}

// Index contains CHM index data
type Index struct {
	properties map[string]string
	root       *IndexItem
}

// NewIndex creates new Index
func NewIndex() *Index {
	return &Index{
		properties: make(map[string]string),
		root:       NewIndexItem("", nil),
	}
}

// Properties returns the properties
func (i *Index) Properties() map[string]string {
	return i.properties
}

// Root returns the root node
func (i *Index) Root() *IndexItem {
	return i.root
}

// SetProp set index property
func (i *Index) SetProp(k, v string) {
	i.properties[k] = v
}

// GetProp returns index property
func (i *Index) GetProp(k string) string {
	return i.properties[k]
}

// Serialize serializes the index
func (i *Index) Serialize(b *Buffer) {
	b.Line(`<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML//EN">`)
	b.Line("<HTML>")
	b.Line("<HEAD>")
	b.Line(`<meta name="GENERATOR" content="Microsoft&reg; HTML Help Workshop 4.1">`)
	b.Line("<!-- Sitemap 1.0 -->")
	b.Line("</HEAD><BODY>")

	// Write properties
	if len(i.properties) > 0 {
		b.Indent(`<OBJECT type="text/site properties">`)
		for k, v := range i.properties {
			if v != "" {
				b.Line(`<param name="%s" value="%s">`, k, v)
			}
		}
		b.Unindent("</OBJECT>")
	}
	i.Root().Serialize(b)
	b.Line("</BODY></HTML>")
}

// IndexItem represents a keyword in the index
type IndexItem struct {
	keyword  string
	locals   []*Local
	children []*IndexItem
	childMap map[string]*IndexItem
	parent   *IndexItem
}

// NewIndexItem creates a IndexItem
func NewIndexItem(keyword string, parent *IndexItem) *IndexItem {
	return &IndexItem{
		keyword:  keyword,
		locals:   make([]*Local, 0, 1),
		children: make([]*IndexItem, 0),
		childMap: make(map[string]*IndexItem),
		parent:   parent,
	}
}

// Add adds subkeyword to the keyword
func (i *IndexItem) Add(keyword string) *IndexItem {
	if v, ok := i.childMap[keyword]; ok {
		return v
	}
	v := NewIndexItem(keyword, i)
	i.childMap[keyword] = v
	i.children = append(i.children, v)
	return v
}

// AddLocal adds a new local to the keyword
func (i *IndexItem) AddLocal(href, title string) *Local {
	href = strings.TrimSpace(href)
	title = strings.TrimSpace(title)
	for _, v := range i.locals {
		if v.href == href && v.title == title {
			return v
		}
	}
	l := Local{href, title}
	i.locals = append(i.locals, &l)
	return &l
}

// IsRoot returns true if this is the root node
func (i *IndexItem) IsRoot() bool {
	return i.parent == nil
}

// Sort sorts the index subkeywords
func (i *IndexItem) Sort() {
	sort.Sort(IndexSorter(i.children))
}

// Serialize creates the .hhk content
func (i *IndexItem) Serialize(b *Buffer) {
	b.Indent(`<LI> <OBJECT type="text/sitemap">`)
	b.Line(`<param name="Name" value="%s">`, strings.TrimSpace(i.keyword))

	sort.Sort(LocalSorter(i.locals))

	for _, l := range i.locals {
		if l.title != "" {
			b.Line(`<param name="Name" value="%s">`, strings.TrimSpace(l.title))
		}
		b.Line(`<param name="Local" value="%s">`, l.href)
	}
	b.Unindent("</OBJECT>")

	if len(i.children) > 0 {
		i.Sort()
		b.Indent("<UL>")
		for _, c := range i.children {
			c.Serialize(b)
		}
		b.Unindent("</UL>")
	}
}

// LocalSorter sorts the keywords
type LocalSorter []*Local

func (a LocalSorter) Len() int      { return len(a) }
func (a LocalSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LocalSorter) Less(i, j int) bool {
	x, y := a[i], a[j]
	switch {
	case x.title == "" && y.title != "":
		return true
	case x.title != "" && y.title == "":
		return false
	case x.title != "" && y.title != "":
		return strings.ToLower(x.title) < strings.ToLower(y.title)
	default:
		return strings.ToLower(x.href) < strings.ToLower(y.href)
	}
}

// IndexSorter sorts the keywords
type IndexSorter []*IndexItem

var nameRe = regexp.MustCompile(`^\w+`)
var pkgRe = regexp.MustCompile(`\((const|var|func|type) in (.+)\)$`)
var methodRe = regexp.MustCompile(`\((method) of (\w+?) in (.+)\)$`)

var typeWeights = map[string]int{
	"const":  1,
	"var":    2,
	"func":   3,
	"type":   4,
	"method": 5,
}

// splitKeyword splits the index keywork into package, struct, name
func splitKeyword(keyword string) (name, typeName, structName, packageName string) {
	name = nameRe.FindString(keyword)
	if name == "" {
		log.Fatal("name is empty in", keyword)
	}

	if matches := methodRe.FindStringSubmatch(keyword); matches != nil {
		typeName, structName, packageName = matches[1], matches[2], matches[3]
	} else if matches := pkgRe.FindStringSubmatch(keyword); matches != nil {
		typeName, packageName = matches[1], matches[2]
	}
	return
}

func comparePackage(p1, p2 string) int {
	p1p := strings.Split(strings.ToLower(p1), "/")
	p2p := strings.Split(strings.ToLower(p2), "/")

	if len(p1p) < len(p2p) {
		return -1
	} else if len(p1p) > len(p2p) {
		return 1
	}

	for i := 0; i < len(p1p); i++ {
		v1 := p1p[i]
		v2 := p2p[i]
		if v1 < v2 {
			return -1
		} else if v1 > v2 {
			return 1
		}
	}

	return 0
}

func (a IndexSorter) Len() int      { return len(a) }
func (a IndexSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a IndexSorter) Less(i, j int) bool {
	k1 := strings.ToLower(a[i].keyword)
	n1, t1, s1, p1 := splitKeyword(k1)
	k2 := strings.ToLower(a[j].keyword)
	n2, t2, s2, p2 := splitKeyword(k2)
	pv := comparePackage(p1, p2)
	if n1 != n2 {
		return n1 < n2
	} else if pv != 0 {
		return pv < 0
	} else if t1 != t2 {
		return typeWeights[t1] < typeWeights[t2]
	} else if s1 != s2 {
		return s1 < s2
	}
	return k1 < k2
}
