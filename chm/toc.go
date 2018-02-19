package chm

import (
	"log"
	"sort"
	"strings"
)

// Toc contains the table of contents
type Toc struct {
	properties map[string]string
	root       *TocItem
}

// NewToc creates new Toc
func NewToc() *Toc {
	return &Toc{
		properties: make(map[string]string),
		root:       NewTocItem("", "", nil),
	}
}

// GetProperties returns the properties
func (t *Toc) Properties() map[string]string {
	return t.properties
}

// GetRoot returns the root node
func (t *Toc) Root() *TocItem {
	return t.root
}

func (t *Toc) SetProp(k, v string) {
	t.properties[k] = v
}

// Serialize serializes the toc
func (t *Toc) Serialize(b *Buffer) {
	b.Line(`<!DOCTYPE HTML PUBLIC "-//IETF//DTD HTML//EN">`)
	b.Line("<HTML>")
	b.Line("<HEAD>")
	b.Line(`<meta name="GENERATOR" content="Microsoft&reg; HTML Help Workshop 4.1">`)
	b.Line("<!-- Sitemap 1.0 -->")
	b.Line("</HEAD><BODY>")

	// Write properties
	if len(t.properties) > 0 {
		b.Indent(`<OBJECT type="text/site properties">`)
		for k, v := range t.properties {
			if v != "" {
				b.Line(`<param name="%s" value="%s">`, k, v)
			}
		}
		b.Unindent("</OBJECT>")
	}
	t.Root().Serialize(b)
	b.Line("</BODY></HTML>")
}

// TocItem is an item on the Toc
type TocItem struct {
	label    string
	href     string
	children []*TocItem
	parent   *TocItem
	image    int
}

// NewTocItem creates new TocItem
func NewTocItem(label, href string, parent *TocItem) *TocItem {
	return &TocItem{
		label:    label,
		href:     href,
		children: make([]*TocItem, 0),
		parent:   parent,
	}
}

func (t *TocItem) Label() string {
	return t.label
}

// Add adds a new child toc item
func (t *TocItem) Add(label, href string) *TocItem {
	label = strings.TrimSpace(label)
	href = strings.TrimSpace(href)
	for _, c := range t.children {
		if c.label == label && c.href == href {
			return c
		}
	}
	c := NewTocItem(label, href, t)
	t.children = append(t.children, c)
	return c
}

// Parent returns parent
func (t *TocItem) Parent() *TocItem {
	return t.parent
}

// IsRoot returns true if this is the root node
func (t *TocItem) IsRoot() bool {
	return t.parent == nil
}

// Level returns the item level
func (t *TocItem) Level() int {
	l := 0
	p := t.Parent()
	for p != nil {
		l++
		p = p.Parent()
	}
	return l
}

// TagAs sets the item image
func (t *TocItem) TagAs(tag string) {
	switch tag {
	case "folder", "directory":
		t.image = 5
	case "file":
		t.image = 11
	case "book", "heading":
		// use default icon
	case "function":
		t.image = 17
	case "method":
		t.image = 19
	case "field":
		t.image = 35
	case "type", "class", "interface":
		t.image = 37
	default:
		log.Fatal("Unknown tag: ", t)
	}
}

func (t *TocItem) Sort() {
	sort.Sort(TocSorter(t.children))
}

// Serialize serializes the toc
func (t *TocItem) Serialize(b *Buffer) {
	if !t.IsRoot() {
		b.Indent(`<LI> <OBJECT type="text/sitemap">`)
		b.Line(`<param name="Name" value="%s">`, t.label)
		if t.href != "" {
			b.Line(`<param name="Local" value="%s">`, t.href)
		}
		if t.image > 0 {
			b.Line(`<param name="ImageNumber" value="%d">`, t.image)
		} else {
			if t.href == "" {
				// set folder icon for topic without href
				b.Line(`<param name="ImageNumber" value="5">`)
			}
		}
		b.Unindent("</OBJECT>")
	}
	if len(t.children) > 0 {
		b.Indent("<UL>")
		for _, c := range t.children {
			c.Serialize(b)
		}
		b.Unindent("</UL>")
	}
}

// TocSorter sorts the items
type TocSorter []*TocItem

func (a TocSorter) Len() int      { return len(a) }
func (a TocSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TocSorter) Less(i, j int) bool {
	return strings.ToLower(a[i].label) < strings.ToLower(a[j].label)
}
