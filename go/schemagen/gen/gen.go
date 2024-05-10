package gen

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type Gen struct {
	SourceDir string
	SourceExt string
	DestDir   string
}

func NewGen(sourceDir string, destDir string) *Gen {
	return &Gen{
		SourceDir: sourceDir,
		SourceExt: ".go",
		DestDir:   destDir,
	}
}

func (g *Gen) Generate() error {
	var err error

	var files []*SourceFile
	if files, err = readFiles(g.SourceDir, g.SourceExt); err != nil {
		return err
	}

	for _, f := range files {
		err = g.genSchemaFromFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Gen) genSchemaFromFile(source *SourceFile) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "./message/"+source.SourcePath(), nil, parser.ParseComments)
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
			fmt.Printf("SKIP %s GenDecl.Doc is nil\n", source.SourcePath())
			continue
		}

		var targetStruct *schemaStruct
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

			targetStruct = newSchemaStruct()
			targetStruct.Name = currType.Name.Name
			targetStruct.Decl = genD
		}

		if targetStruct == nil {
			continue
		}

		var needCodegen bool
		for _, comment := range genD.Doc.List {
			needCodegen = needCodegen || strings.HasPrefix(comment.Text, "// schemagen:")
			if len(comment.Text) > 13 {
				targetStruct.Name = strings.TrimSpace(strings.Replace(comment.Text, "// schemagen:", "", 1))
			}
		}

		if !needCodegen {
			continue
		}

		if err = targetStruct.Parse(); err != nil {
			return err
		}

		out, err := os.Create("../schema/" + source.DestPath(".json"))
		if err != nil {
			log.Fatal(err.Error())
		}
		defer out.Close()

		j, err := json.Marshal(targetStruct.Schema)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Fprintln(out, string(j))

	}

	fmt.Println(node.Name.Name)

	return nil
}
