package gen

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/go-faster/errors"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

//go:embed _template/*.tmpl
var templates embed.FS

// Funcs returns functions which used in templates.
func Funcs() template.FuncMap {
	return template.FuncMap{
		"trim":       strings.TrimSpace,
		"lower":      strings.ToLower,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"contains":   strings.Contains,
	}
}

// Template parses and returns vendored code generation templates.
func Template() *template.Template {
	tmpl := template.New("templates").Funcs(Funcs())
	tmpl = template.Must(tmpl.ParseFS(templates, "_template/*.tmpl"))
	return tmpl
}

// config is input data for templates.
type config struct {
	Package string
	Structs []structDef
}

type structDef struct {
	// Name of struct, just like that: `type Name struct {}`.
	Name string
	// Receiver name. E.g. "m" for Message.
	Receiver string

	Fields []fieldDef
}

type Generator struct {
	pkg *packages.Package

	// structs definitions.
	structs []structDef
}

type fieldDef struct {
	Name         string
	Key          string
	EncodeMethod string
	DecodeMethod string
}

// WriteSource writes generated definitions to fs.
func (g *Generator) WriteSource(fs FileSystem, pkgName string, t *template.Template) error {
	w := &writer{
		pkg:   pkgName,
		fs:    fs,
		t:     t,
		buf:   new(bytes.Buffer),
		wrote: map[string]bool{},
	}
	if err := w.WriteStructs(g.structs); err != nil {
		return errors.Wrap(err, "structs")
	}

	return nil
}

func (g *Generator) encodeMethod(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.String:
			return "StrEscape"
		case types.Int:
			return "Int"
		default:
			panic("unhandled default case")
		}
	}
	return "Encode"
}

func (g *Generator) decodeMethod(t types.Type) string {
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.String:
			return "Str"
		case types.Int:
			return "Int"
		default:
			panic("unhandled default case")
		}
	}
	return "Decode"
}

func (g *Generator) makeStructs() error {
	i := inspector.New(g.pkg.Syntax)
	var toResolve []string
	i.Nodes([]ast.Node{
		&ast.GenDecl{},
	}, func(node ast.Node, push bool) (proceed bool) {
		genDecl := node.(*ast.GenDecl)
		if genDecl.Doc == nil {
			return false
		}
		typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			return false
		}
		for _, comment := range genDecl.Doc.List {
			switch comment.Text {
			case "//jxgen:json":
				fmt.Println("found entity", typeSpec.Name.Name)
				toResolve = append(toResolve, typeSpec.Name.Name)
			}
		}
		return false
	})
	for _, name := range toResolve {
		s := g.pkg.Types.Scope().Lookup(name).Type().Underlying().(*types.Struct)
		var fields []fieldDef
		for i := 0; i < s.NumFields(); i++ {
			field := s.Field(i)
			tag := reflect.StructTag(s.Tag(i))
			fmt.Printf("Field: %s, Type: %s, Tag: %s\n", field.Name(), field.Type(), tag.Get("json"))
			fields = append(fields, fieldDef{
				Name:         field.Name(),
				Key:          tag.Get("json"),
				EncodeMethod: g.encodeMethod(field.Type()),
				DecodeMethod: g.decodeMethod(field.Type()),
			})
		}
		fmt.Println("struct", name)
		g.structs = append(g.structs, structDef{
			Name:     name,
			Receiver: strings.ToLower(name[:1]),
			Fields:   fields,
		})
	}
	return nil
}

func NewGenerator(pkg *packages.Package) (*Generator, error) {
	g := &Generator{
		pkg: pkg,
	}
	if pkg.Name == "" {
		return nil, errors.New("package name is empty")
	}
	if err := g.makeStructs(); err != nil {
		return nil, errors.Wrap(err, "make structs")
	}

	return g, nil
}

// FileSystem represents a directory of generated package.
type FileSystem interface {
	WriteFile(baseName string, source []byte) error
}
type writer struct {
	pkg   string
	fs    FileSystem
	t     *template.Template
	buf   *bytes.Buffer
	wrote map[string]bool
}

// Generate executes template to file using config.
func (w *writer) Generate(templateName, fileName string, cfg config) error {
	if cfg.Package == "" {
		cfg.Package = w.pkg
	}
	if w.wrote[fileName] {
		return errors.Errorf("name collision (already wrote %s)", fileName)
	}

	w.buf.Reset()
	if err := w.t.ExecuteTemplate(w.buf, templateName, cfg); err != nil {
		return errors.Wrapf(err, "execute template %s for %s", templateName, fileName)
	}
	if err := w.fs.WriteFile(fileName, w.buf.Bytes()); err != nil {
		_ = os.WriteFile(fileName+".dump", w.buf.Bytes(), 0600)
		return errors.Wrapf(err, "write file %s", fileName)
	}
	w.wrote[fileName] = true

	return nil
}

// WriteStructs writes structure definitions to corresponding files.
func (w *writer) WriteStructs(structs []structDef) error {
	cfg := config{
		Package: w.pkg,
		Structs: structs,
	}
	name := "jx_gen.go"
	if err := w.write(name, cfg); err != nil {
		return errors.Wrap(err, "write")
	}

	return nil
}

func (w *writer) write(fileName string, cfg config) error {
	if err := w.Generate("main", fileName, cfg); err != nil {
		return err
	}

	return nil
}
