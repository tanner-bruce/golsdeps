package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("USAGE: golsdeps package file [ file ... ]")
		os.Exit(1)
	}
	if err := run(os.Args[1], os.Args[2:]...); err != nil {
		fmt.Println(err)
	}
}

func packageForFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	rd := bufio.NewReader(f)
	var pkg string
	for pkg == "" {
		s, err := rd.ReadString('\n')
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(s, "package") {
			continue
		}
		ss := strings.Split(s, " ")
		if ss[0] != "package" {
			continue
		}
		pkg = strings.Trim(ss[1], "\n")
	}
	f.Close()
	basePkg := filepath.Dir(filepath.Dir(filename))
	return basePkg + "/" + pkg, nil
}

func run(pkg string, modifiedFiles ...string) error {
	modifiedPackages := map[string]bool{}
	basepkg, err := packages.Load(nil, ".")
	if err != nil {
		return err
	}
	bp := basepkg[0]
	for _, f := range modifiedFiles {
		pkg, err := packageForFile(f)
		if err != nil {
			return err
		}
		modifiedPackages[bp.PkgPath+"/"+pkg] = true
	}
	packagesToTest := map[string]bool{}

	ps, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedImports | packages.NeedDeps,
		Tests: true,
	}, pkg)
	if err != nil {
		return err
	}
	for _, p := range ps {
		for pk := range p.Imports {
			if modifiedPackages[pk] {
				packagesToTest[p.PkgPath] = true
			}
		}
	}
	for k := range packagesToTest {
		fmt.Println(k)
	}
	return nil
}
