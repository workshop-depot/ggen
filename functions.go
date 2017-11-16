package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

//-----------------------------------------------------------------------------

func process(wd string) {
	if err := activate(newProcessor(wd)); err != nil {
		logerr.Printf("%+v", err)
	}
}

//-----------------------------------------------------------------------------

func extractBlankImports(fpath string) ([]string, error) {
	fset := token.NewFileSet()
	fast, err := parser.ParseFile(fset, fpath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var blankImports []string
	for _, vi := range fast.Imports {
		if vi.Path == nil || vi.Name == nil {
			continue
		}
		pkg := vi.Path.Value
		pkg = strings.Trim(pkg, `"`)
		if vi.Name.Name != "_" {
			continue
		}
		blankImports = append(blankImports, pkg)
	}
	return blankImports, nil
}

//-----------------------------------------------------------------------------

func checkFileSet(wd string) bool {
	list, err := ioutil.ReadDir(wd)
	if err != nil {
		panic(err)
	}
	paramExist := false
	defgnExist := false
	speclExist := false
	for _, vf := range list {
		if vf.IsDir() {
			continue
		}
		if vf.Name() == paramFN {
			paramExist = true
		}
		if vf.Name() == defgnFN {
			defgnExist = true
		}
		if vf.Name() == speclFN {
			speclExist = true
		}
	}
	return paramExist && defgnExist && speclExist
}

//-----------------------------------------------------------------------------

func checkSrcDir() (string, error) {
	gopath := os.Getenv("GOPATH")
	if err := dirExists(gopath); err != nil {
		return "", errors.WithMessage(err, fmt.Sprintf("not found, $GOPATH = %v", gopath))
	}

	parts := strings.Split(gopath, string([]rune{filepath.ListSeparator}))
	gopath = parts[0]

	src := filepath.Join(gopath, "src")
	if err := dirExists(src); err != nil {
		return "", errors.WithMessage(err, "src directory not found")
	}

	return src, nil
}

//-----------------------------------------------------------------------------

func mkdir(d string) error {
	return os.Mkdir(d, 0777)
}

//-----------------------------------------------------------------------------

func writeFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0777)
}

//-----------------------------------------------------------------------------

func fileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errNotFile
	}
	return nil
}

//-----------------------------------------------------------------------------

func cp(src, dst string, overwrite ...bool) (funcErr error) {
	fw := false
	if len(overwrite) > 0 {
		fw = overwrite[0]
	}
	exists := true
	if _, err := os.Stat(dst); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		exists = false
	}
	if exists && !fw {
		return nil
	}
	fsrc, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fsrc.Close()
	fdst, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		err := fdst.Close()
		if funcErr != nil {
			return
		}
		funcErr = err
	}()
	if _, err := io.Copy(fdst, fsrc); err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------

func dirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errNotDir
	}
	return nil
}

//-----------------------------------------------------------------------------

func errorf(format string, a ...interface{}) error {
	return sentinelErr(fmt.Sprintf(format, a...))
}

//-----------------------------------------------------------------------------

type sentinelErr string

func (v sentinelErr) Error() string { return string(v) }

//-----------------------------------------------------------------------------
