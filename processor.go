package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type processor struct {
	wd    string
	gosrc string

	defgnBlankImports []string
	impgnBlankImports []string
}

func newProcessor(wd string) *processor {
	return &processor{
		wd: wd,
	}
}

func (pc *processor) Activate() (state, error) {
	return stateFunc(pc.extractBlankImports), nil
}

func (pc *processor) extractBlankImports() (state, error) {
	files, err := ioutil.ReadDir(pc.wd)
	if err != nil {
		return nil, err
	}
	for _, vf := range files {
		if vf.IsDir() {
			continue
		}
		name := vf.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, paramFN) ||
			strings.HasSuffix(name, defgnFN) ||
			strings.HasSuffix(name, speclFN) {
			if name == defgnFN {
				imports, err := extractBlankImports(filepath.Join(pc.wd, name))
				if err != nil {
					return nil, err
				}
				pc.defgnBlankImports = append(pc.defgnBlankImports, imports...)
			}
		} else {
			imports, err := extractBlankImports(filepath.Join(pc.wd, name))
			if err != nil {
				return nil, err
			}
			pc.impgnBlankImports = append(pc.impgnBlankImports, imports...)
		}
	}
	return stateFunc(pc.sync), nil
}

func (pc *processor) sync() (state, error) {
	src, err := checkSrcDir()
	if err != nil {
		return nil, err
	}
	pc.gosrc = src
	sort.Strings(pc.defgnBlankImports)
	sort.Strings(pc.impgnBlankImports)
	if len(pc.impgnBlankImports) > 0 {
		return stateFunc(pc.syncImpgn), nil
	}
	return stateFunc(pc.syncDefgn), nil
}

func (pc *processor) syncImpgn() (state, error) {
	if len(pc.impgnBlankImports) > 1 {
		logwrn.Println("more than one blank imports found. only the first found generic definition will be used.")
	}
	// find first generic definition (from sorted list of blank imports)
	var first string
	for _, vg := range pc.impgnBlankImports {
		path := filepath.Join(pc.gosrc, vg)
		if dirExists(path) != nil {
			continue
		}
		if checkFileSet(path) {
			first = path
			break
		}
	}
	if first == "" {
		return nil, errorf("none of blank imports are a generic definition.")
	}
	// copy generic definition to generic implementation
	files, err := ioutil.ReadDir(first)
	if err != nil {
		return nil, err
	}
	for _, vf := range files {
		if vf.IsDir() {
			continue
		}
		name := vf.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		force := false
		if strings.HasSuffix(name, defgnFN) {
			force = true
		}
		if err := cp(filepath.Join(first, name), filepath.Join(pc.wd, name), force); err != nil {
			return nil, err
		}
	}
	// adopt package name
	files, err = ioutil.ReadDir(pc.wd)
	if err != nil {
		return nil, err
	}
	var pkgName string
	for _, vf := range files {
		if vf.IsDir() {
			continue
		}
		name := vf.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, paramFN) ||
			strings.HasSuffix(name, defgnFN) ||
			strings.HasSuffix(name, speclFN) {
			continue
		}
		sample := filepath.Join(pc.wd, name)
		fset := token.NewFileSet()
		fast, err := parser.ParseFile(fset, sample, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		pkgName = fmt.Sprintf("%v", fast.Name)
		break
	}
	for _, vf := range files {
		if vf.IsDir() {
			continue
		}
		name := vf.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if strings.HasSuffix(name, paramFN) ||
			strings.HasSuffix(name, defgnFN) ||
			strings.HasSuffix(name, speclFN) {
			p := filepath.Join(pc.wd, name)
			fset := token.NewFileSet()
			fast, err := parser.ParseFile(fset, p, nil, parser.PackageClauseOnly)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			start := int(fset.Position(fast.Pos()).Offset)
			end := int(fset.Position(fast.End()).Offset)
			content, err := ioutil.ReadFile(p)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			if err := writeFile(p, fmt.Sprintf("%s%s%s", content[:start], "package "+pkgName, content[end:])); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

var qq int

func (pc *processor) syncDefgn() (state, error) {
	if !checkFileSet(pc.wd) {
		return nil, errorf("some files are missing (this package apparently is not a generic implementation; for a generic definition %v, %v and %v should exist)", defgnFN, paramFN, speclFN)
	}
	files, err := ioutil.ReadDir(pc.wd)
	if err != nil {
		return nil, err
	}
	var dstPkg string
	for _, vf := range files {
		if vf.IsDir() {
			continue
		}
		name := vf.Name()
		if filepath.Ext(name) != ".go" {
			continue
		}
		if name != defgnFN {
			continue
		}
		sample := filepath.Join(pc.wd, name)
		fset := token.NewFileSet()
		fast, err := parser.ParseFile(fset, sample, nil, parser.PackageClauseOnly)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		dstPkg = fmt.Sprintf("%v", fast.Name)
		break
	}
	for _, vg := range pc.defgnBlankImports {
		path := filepath.Join(pc.gosrc, vg)
		if dirExists(path) != nil {
			continue
		}
		if !checkFileSet(path) {
			continue
		}
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		// find imported generic definition package name
		var srcPkg string
		for _, vf := range files {
			if vf.IsDir() {
				continue
			}
			name := vf.Name()
			if name != defgnFN {
				continue
			}
			sample := filepath.Join(path, name)
			fset := token.NewFileSet()
			fast, err := parser.ParseFile(fset, sample, nil, parser.PackageClauseOnly)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			srcPkg = fmt.Sprintf("%v", fast.Name)
			break
		}
		// copy generic definition files
		for _, vf := range files {
			if vf.IsDir() {
				continue
			}
			name := vf.Name()
			if filepath.Ext(name) != ".go" {
				continue
			}
			force := false
			if strings.HasSuffix(name, defgnFN) {
				force = true
			}
			dstName := name
			if name == paramFN ||
				name == defgnFN ||
				name == speclFN {
				dstName = srcPkg + "-" + name
			}
			s := filepath.Join(path, name)
			d := filepath.Join(pc.wd, dstName)
			if err := cp(s, d, force); err != nil {
				return nil, err
			}
		}
		files, err = ioutil.ReadDir(pc.wd)
		if err != nil {
			return nil, err
		}
		for _, vf := range files {
			if vf.IsDir() {
				continue
			}
			name := vf.Name()
			if filepath.Ext(name) != ".go" {
				continue
			}
			if strings.HasSuffix(name, paramFN) ||
				strings.HasSuffix(name, defgnFN) ||
				strings.HasSuffix(name, speclFN) {
				p := filepath.Join(pc.wd, name)
				fset := token.NewFileSet()
				fast, err := parser.ParseFile(fset, p, nil, parser.PackageClauseOnly)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				if fmt.Sprintf("%v", fast.Name) == dstPkg {
					continue
				}
				start := int(fset.Position(fast.Pos()).Offset)
				end := int(fset.Position(fast.End()).Offset)
				content, err := ioutil.ReadFile(p)
				if err != nil {
					return nil, errors.WithStack(err)
				}
				if err := writeFile(
					filepath.Join(pc.wd, name),
					fmt.Sprintf("%s%s%s", content[:start], "package "+dstPkg, content[end:])); err != nil {
					return nil, err
				}
			}
		}
	}
	return nil, nil
}
