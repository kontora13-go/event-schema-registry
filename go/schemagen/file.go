package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"strings"
)

type sourceFile struct {
	name string
	path string
}

func newSourceFile(path string) *sourceFile {
	p := strings.Split(path, "/")

	f := &sourceFile{}
	f.name, _ = strings.CutSuffix(p[len(p)-1], ".go")
	for i := 0; i < len(p)-1; i++ {
		f.path = f.path + p[i] + "/"
	}

	return f
}

func readFiles(dir string) ([]*sourceFile, error) {
	fileSystem := os.DirFS(dir)

	files := make([]*sourceFile, 0)

	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if path == "." {
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		f := newSourceFile(path)
		files = append(files, f)

		return nil
	})

	return files, err
}

func (f *sourceFile) sourcePath() string {
	return f.path + f.name + ".go"
}

func (f *sourceFile) destPath() string {
	return f.path + f.name + ".json"
}

func (f *sourceFile) genFromFile() error {
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, f.sourcePath(), nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, gd := range node.Decls {
		genD, ok := gd.(*ast.GenDecl)
		if !ok {
			fmt.Printf("SKIP %T is not *ast.GenDecl\n", gd)
			continue
		}

		if genD.Doc == nil {
			fmt.Printf("SKIP %s GenDecl.Doc is nil\n", f.sourcePath())
			continue
		}

		targetStruct := &schemaStruct{}
		var thisIsStruct bool
		for _, spec := range genD.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				fmt.Printf("SKIP %T is not ast.TypeSpec\n", spec)
				continue
			}

			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				fmt.Printf("SKIP %T is not ast.StructType\n", currStruct)
				continue
			}
			targetStruct.Name = currType.Name.Name
			thisIsStruct = true
		}

		var needCodegen bool
		if thisIsStruct {
			for _, comment := range genD.Doc.List {
				needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// schemagen:")
				if len(comment.Text) > 13 {
					targetStruct.Name = strings.TrimSpace(strings.Replace(comment.Text, "// schemagen:", "", 1))
				}
			}
			targetStruct.Decl = genD
		}

		if needCodegen {
			//targetStruct.generateSchema(f)
		}
	}

	fmt.Println(node.Name.Name)
	//funcGetResponse.Execute(out, tpl{node.Name.Name})

	return nil
}
