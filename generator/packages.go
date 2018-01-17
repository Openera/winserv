// Copyright (c) 2018 LiveQoS. All rights reserved.
//

package main

import (
	"bufio"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

const topDir = "p"

var pkgNames = []string{"a",
	"b",
	"c",
	"d",
	"e"}

func main() {
	fmt.Printf("Create\n")
	os.RemoveAll(topDir)
	os.MkdirAll(topDir, os.FileMode(0755))
	for _, pkg := range pkgNames {
		path := filepath.Join(topDir, pkg)
		fmt.Printf("Create package %v\n", path)
		os.MkdirAll(path, os.FileMode(0755))
		createPackageSource(pkg, path)
	}
}

func createPackageSource(pkgName string, pkgPath string) (err error) {

	filename := "code.go"
	file, err := os.Create(filepath.Join(pkgPath, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	wr := bufio.NewWriter(file)

	const tmpl = `// Generated file. Do not modify by hand.

package {{ .Name }}

import (
	"github.com/Openera/winserv/out"
)

var variable = out.LogString("{{ .Name }}/var")`

	t := template.Must(template.New("tmpl").Parse(tmpl))
	t.Execute(wr, struct {
		Name string
	}{pkgName})

	wr.Flush()

	return err
}
