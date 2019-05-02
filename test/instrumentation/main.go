/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"k8s.io/kubernetes/third_party/go-srcimporter"
)

type visitor struct {
	FileSet      *token.FileSet
	cMap         ast.CommentMap
	Rparen *token.Pos
	Lparen *token.Pos
	Fname string
	keyValues map[string]string
	collected []opts
}

func newVisitor() *visitor {
	return &visitor{
		FileSet: token.NewFileSet(),
		collected: make([]opts,0),
		keyValues: make(map[string]string, 0),
	}
}

type opts struct {
	fname string
	keyValues map[string]string
}


type analyzer struct {
	fset      *token.FileSet // positions are relative to fset
	conf      types.Config
	ctx       build.Context
	failed    bool
	donePaths map[string]interface{}
	errors    []string
}

func newAnalyzer() *analyzer {
	ctx := build.Default
	ctx.CgoEnabled = true

	a := &analyzer{
		fset:      token.NewFileSet(),
		ctx:       ctx,
		donePaths: make(map[string]interface{}),
	}
	a.conf = types.Config{
		FakeImportC: true,
		Error:       a.handleError,
		Sizes:       types.SizesFor("gc", a.ctx.GOARCH),
	}

	a.conf.Importer = srcimporter.New(
		&a.ctx, a.fset, make(map[string]*types.Package))

	return a
}

func (a *analyzer) handleError(err error) {
	if e, ok := err.(types.Error); ok {
		// useful for some ignores:
		// path := e.Fset.Position(e.Pos).String()
		ignore := false
		// TODO(rmmh): read ignores from a file, so this code can
		// be Kubernetes-agnostic. Unused ignores should be treated as
		// errors, to ensure coverage isn't overly broad.
		if strings.Contains(e.Msg, "GetOpenAPIDefinitions") {
			// TODO(rmmh): figure out why this happens.
			// cmd/kube-apiserver/app/server.go:392:70
			// test/integration/framework/master_utils.go:131:84
			ignore = true
		}
		if ignore {
			return
		}
	}
	a.errors = append(a.errors, err.Error())
	a.failed = true
}

func (a *analyzer) dumpAndResetErrors() []string {
	es := a.errors
	a.errors = nil
	return es
}

// collect extracts test metadata from a file.
func (a *analyzer) collect(v *visitor, dir string) {

	// don't collect if we've already done so
	if _, ok := a.donePaths[dir]; ok {
		return
	}
	a.donePaths[dir] = nil

	// Create the AST by parsing src.
	fs, err := parser.ParseDir(a.fset, dir, nil, parser.AllErrors)

	//fmt.Printf("alalal AST %v\n\n", fs)
	if err != nil {
		fmt.Println("ERROR(syntax)", err)
		a.failed = true
		return
	}

	for _, p := range fs {
		// returns first error, but a.handleError deals with it
		files := a.filterFiles(p.Files)
		_ = a.check(v, dir, files)
	}
}

// filterFiles restricts a list of files to only those that should be built by
// the current platform. This includes both build suffixes (_windows.go) and build
// tags ("// +build !linux" at the beginning).
func (a *analyzer) filterFiles(fs map[string]*ast.File) []*ast.File {
	files := []*ast.File{}
	for _, f := range fs {
		fpath := a.fset.File(f.Pos()).Name()
		dir, name := filepath.Split(fpath)
		matches, err := a.ctx.MatchFile(dir, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR reading %s: %s\n", fpath, err)
			a.failed = true
			continue
		}
		if matches {
			files = append(files, f)
		}
	}
	return files
}

// walk the AST and gather our data related to our relevant function calls
func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.CallExpr:
		if f, ok := t.Fun.(*ast.Ident); ok {
			v.Fname = f.String()
			// todo: use a regex match to catch all of our function calls
			if v.Fname == "NewCounter" {
				v.Rparen = &t.Rparen
				v.Lparen = &t.Lparen
				fmt.Printf("func: %v R: %v L: %v\n", f.String(), t.Rparen, t.Lparen)
			}

			return v
		}

	case *ast.KeyValueExpr:
		if v.Rparen != nil {
			// check that we are inside the boundaries of the function call we are interested in
			if *v.Rparen > t.Pos() && *v.Lparen < t.Pos() {
				key := fmt.Sprintf("%v", t.Key)
				value := fmt.Sprintf("%v", t.Value)
				if k, ok := t.Value.(*ast.BasicLit); ok {
					if k.Kind == token.STRING {
						v.keyValues[key] = k.Value
					}
				} else {
					v.keyValues[key] = value
				}
			}

			return v
		}

	default:
		if v.Rparen == nil {
			return v
		}
		// todo: let's use a regex for the function names we want to test
		if v.Fname == "NewCounter" {
			if sl, ok := v.keyValues["StabilityLevel"]; ok{
				if sl == "ALPHA" {
					fmt.Printf("RParen :%+v POS %v | %v | %v \n\n", *v.Rparen, t.Pos(), v.Fname, v.keyValues)
					if *v.Rparen > t.Pos() {
						fmt.Println("blah blah blah blah")
						o := opts{fname: v.Fname, keyValues: v.keyValues}
						v.collected = append(v.collected, o)
						v.Fname = ""
						v.Rparen = nil
						v.keyValues = make(map[string]string, 0)
					}

				}
			}
		}
	}
	return v
}

func (a *analyzer) check(v *visitor, dir string, files []*ast.File) error {
	pkg, err := a.conf.Check(dir, a.fset, files, nil)

	if err != nil {
		return err // type error
	}

	// lifted from test/typecheck/main.goo
	for _, imp := range pkg.Imports() {
		if strings.HasPrefix(imp.Path(), "k8s.io/kubernetes/vendor/") {
			vendorPath := imp.Path()[len("k8s.io/kubernetes/"):]

			//fmt.Println("recursively checking vendor path:", vendorPath)
			a.collect(v, vendorPath)
		}
	}
	// let's walk the ast.
	for _, f := range files {
		ast.Walk(v, f)
	}
	// output our collected objects
	for _, c := range v.collected {
		fmt.Printf("COLLECTED: %v %v\n", c.fname, c.keyValues)
	}
	return nil
}

type collector struct {
	dirs []string
}

// lifted from test/typecheck/main.go
func (c *collector) handlePath(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		// Ignore hidden directories (.git, .cache, etc)
		if len(path) > 1 && path[0] == '.' ||
		// Staging code is symlinked from vendor/k8s.io, and uses import
		// paths as if it were inside of vendor/. It fails typechecking
		// inside of staging/, but works when typechecked as part of vendor/.
			path == "staging" ||
		// OS-specific vendor code tends to be imported by OS-specific
		// packages. We recursively typecheck imported vendored packages for
		// each OS, but don't typecheck everything for every OS.
			path == "vendor" ||
			path == "_output" ||
		// This is a weird one. /testdata/ is *mostly* ignored by Go,
		// and this translates to kubernetes/vendor not working.
		// edit/record.go doesn't compile without gopkg.in/yaml.v2
		// in $GOSRC/$GOROOT (both typecheck and the shell script).
			path == "pkg/kubectl/cmd/testdata/edit" {
			return filepath.SkipDir
		}
		c.dirs = append(c.dirs, path)
	}
	return nil
}


func main() {
	flag.Parse()
	args := flag.Args()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "USAGE: %s <DIR or FILE> [...]\n", os.Args[0])
		os.Exit(64)
	}

	if len(args) == 0 {
		args = append(args, "pkg/util/metrics")
	}

	c := collector{}
	for _, arg := range args {
		err := filepath.Walk(arg, c.handlePath)
		if err != nil {
			log.Fatalf("Error walking: %v", err)
		}
	}
	sort.Strings(c.dirs)
	fmt.Printf("%v\n", c.dirs)

	var wg sync.WaitGroup
	wg.Add(1)
	func() {
		a := newAnalyzer()
		for _, dir := range c.dirs {
			a.collect(newVisitor(), dir)
		}
		wg.Done()
	}()


	wg.Wait()
}
