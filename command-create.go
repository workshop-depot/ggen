package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

func cmdCreate(*cli.Context) error {
	conf.Create.Name = strings.TrimSpace(conf.Create.Name)
	if conf.Create.Name == "" {
		logerr.Println("name must be specifiec for the generic package")
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		logerr.Println(err)
		return nil
	}

	pkgName := strings.ToLower(conf.Create.Name)
	typeName := strings.Title(conf.Create.Name)

	gd := filepath.Join(wd, pkgName)
	if err := mkdir(gd); err != nil {
		logerr.Println(err)
		return nil
	}

	paramPath := filepath.Join(gd, paramFN)
	paramContent := fmt.Sprintf(paramFT, pkgName, typeName, typeName)
	if err := writeFile(paramPath, paramContent); err != nil {
		logerr.Println(err)
		return nil
	}
	defgnPath := filepath.Join(gd, defgnFN)
	defgnContent := fmt.Sprintf(defgnFT, pkgName, typeName, typeName, typeName)
	if err := writeFile(defgnPath, defgnContent); err != nil {
		logerr.Println(err)
		return nil
	}
	speclPath := filepath.Join(gd, speclFN)
	speclContent := fmt.Sprintf(speclFT, pkgName)
	if err := writeFile(speclPath, speclContent); err != nil {
		logerr.Println(err)
		return nil
	}

	loginf.Printf("generic code template created at dir= %v", gd)

	return nil
}
