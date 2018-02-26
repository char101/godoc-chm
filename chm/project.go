package chm

import (
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Project contains a project definition
type Project struct {
	name          string
	options       map[string]string
	windowOptions map[string]string
	files         []string
	toc           *Toc
	index         *Index
}

// NewProject creates a Project
func NewProject(name string) *Project {
	p := Project{
		name:  name,
		files: make([]string, 0, 100),
		options: map[string]string{
			"Compatibility":            "1.1 or later",
			"Compiled File":            name + ".chm",
			"Display Compile Progress": "No",
			"Language":                 "0x409 English (United States)",
			"Default Window":           "main",
			"Contents File":            name + ".hhc",
			"Index File":               name + ".hhk",
			"Binary Index":             "No", // with binary index, multi topic keyword will not be displayed
		},
		windowOptions: map[string]string{
			"title": name,
			"id":    "0",
			"navigation_pane_styles": "0x12120",
			"buttons":                "0x10184e",
			"contents_file":          name + ".hhc",
			"index_file":             name + ".hhk",
			"default_topic":          "index.html",
			"home":                   "index.html",
		},
		toc:   NewToc(),
		index: NewIndex(),
	}

	p.toc.SetProp("Window Styles", "0x801627")
	p.toc.SetProp("Font", "Tahoma,8,0")

	p.index.SetProp("Font", "Tahoma,8,0")

	return &p
}

// Name returns name
func (p *Project) Name() string { return p.name }

// Toc returns toc
func (p *Project) Toc() *Toc { return p.toc }

// Index returns index
func (p *Project) Index() *Index { return p.index }

// SetStartFile sets the initial file displayed in the CHM
func (p *Project) SetStartFile(filename string) {
	filename = strings.Replace(filename, "/", "\\", -1)
	p.windowOptions["default_topic"] = filename
	p.windowOptions["home"] = filename
	p.AddFile(filename)
}

// GetCompiledFile returns the compiled file path
func (p *Project) GetCompiledFile() string {
	return p.options["Compiled File"]
}

// SetCompiledFile sets the compiled file path
func (p *Project) SetCompiledFile(path string) {
	p.options["Compiled File"] = path
}

// AddFile adds a file to the project
func (p *Project) AddFile(filename string) {
	filename = strings.Replace(filename, "/", "\\", -1)
	p.files = append(p.files, filename)
}

// Empty returns true of the project has no file
func (p *Project) Empty() bool {
	return len(p.files) == 0
}

// GetFiles returns sorted files with no duplicate
func (p *Project) GetFiles() []string {
	tempFiles := make(map[string]bool)
	for _, v := range p.files {
		tempFiles[v] = true
	}
	tempSlice := make([]string, 0, len(p.files))
	for v := range tempFiles {
		tempSlice = append(tempSlice, v)
	}
	sort.Sort(FileSorter(tempSlice))
	return tempSlice
}

// Serialize serializes the project
func (p *Project) Serialize(b *Buffer) {
	b.Line("[OPTIONS]")
	for k, v := range p.options {
		if v != "" {
			b.Line("%s=%s", k, v)
		}
	}
	b.Line()

	b.Line("[WINDOWS]")

	b.Line("main=%s", p.windowStr())
	b.Line()
	b.Line()

	if len(p.files) > 0 {
		added := make(map[string]bool)
		b.Line("[FILES]")
		for _, f := range p.files {
			f = strings.Replace(f, "/", "\\", -1)
			if _, ok := added[f]; !ok {
				b.Line(f)
			}
			added[f] = true
		}
		b.Line()
	}

	b.Line("[INFOTYPES]")
	b.Line()
}

func (p *Project) windowStr() string {
	numericValue := regexp.MustCompile(`^\d+|0x[0-9a-e]+$`)

	options := []string{
		"title",
		"contents_file",
		"index_file",
		"default_topic",
		"home",
		"jump1",
		"jump1_text",
		"jump2",
		"jump2_text",
		"navigation_pane_styles",
		"navigation_pane_width",
		"buttons",
		"initial_position",
		"style_flags",
		"extended_style_flags",
		"window_show_state",
		"navigation_pane_closed",
		"default_navigation_pane",
		"navigation_pane_position",
		"id",
	}
	values := make([]string, 0, len(options))
	for _, k := range options {
		v, _ := p.windowOptions[k]
		if !(v == "" || numericValue.MatchString(v)) {
			v = `"` + v + `"`
		}
		values = append(values, v)
	}
	return strings.Join(values, ",")
}

// Save saves the project into a file
func (p *Project) Save() {
	Save(p, p.name+".hhp")
	Save(p.toc, p.name+".hhc")
	Save(p.index, p.name+".hhk")
}

// Open opens the project in HTML Help Workshop
func (p *Project) Open() error {
	c := exec.Command(`"C:\Program Files (x86)\HTML Help Workshop\hhw.exe"`, p.name+".hhp")
	return c.Run()
}

// MustOpen opens the project in HTML Help Workshop
func (p *Project) MustOpen() {
	if err := p.Open(); err != nil {
		log.Fatal(err)
	}
}

// Compile compiles the project
func (p *Project) Compile() error {
	c := exec.Command(`C:\Program Files (x86)\HTML Help Workshop\hhc.exe`, p.name+".hhp")
	stdout, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return err
	}
	if err := c.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		return err
	}
	if _, err := io.Copy(os.Stderr, stderr); err != nil {
		return err
	}
	return c.Wait()
}

// MustCompile compies the project
func (p *Project) MustCompile() {
	if err := p.Compile(); err != nil {
		log.Fatalf("%v (%T)", err, err)
	}
}

// FileSorter sorts the items
type FileSorter []string

func (a FileSorter) Len() int      { return len(a) }
func (a FileSorter) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a FileSorter) Less(i, j int) bool {
	return strings.ToLower(a[i]) < strings.ToLower(a[j])
}
