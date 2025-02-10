package main

import (
	"fmt"
	"go/format"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"
	"golang.org/x/tools/go/packages"

	"github.com/ernado/jxgen/internal/gen"
)

type formattedSource struct {
	Format bool
	Root   string
}

func (t formattedSource) WriteFile(name string, content []byte) error {
	out := content
	if t.Format {
		buf, err := format.Source(content)
		if err != nil {
			return err
		}
		out = buf
	}
	return os.WriteFile(filepath.Join(t.Root, name), out, 0600)
}

func run() error {
	pkg := loadPackage(".")
	g, err := gen.NewGenerator(pkg)
	if err != nil {
		return errors.Wrap(err, "create generator")
	}

	fs := formattedSource{
		Root:   ".",
		Format: true,
	}
	if err := g.WriteSource(fs, pkg.Name, gen.Template()); err != nil {
		return errors.Wrap(err, "generate")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		os.Exit(1)
	}
}

func loadPackage(path string) *packages.Package {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports | packages.NeedSyntax | packages.NeedName}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		failErr(fmt.Errorf("loading packages for inspection: %v", err))
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func failErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
