package gen

import (
	"encoding/json"
	"fmt"
	"github.com/kontora13-go/event-schema-registry/schemagen/schema"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

type Gen struct {
	SourceDir     string
	SourceExt     string
	DestDir       string
	MessageStruct *schemaStruct
	GenStruct     []*schemaStruct
	RefStruct     map[string]*schemaStruct
}

func NewGen(sourceDir string, destDir string) *Gen {
	return &Gen{
		SourceDir: sourceDir,
		SourceExt: ".go",
		DestDir:   destDir,
		GenStruct: make([]*schemaStruct, 0),
		RefStruct: make(map[string]*schemaStruct),
	}
}

func (g *Gen) Generate() error {
	var err error

	var files []*SourceFile
	if files, err = readFiles(g.SourceDir, g.SourceExt); err != nil {
		return err
	}

	for _, f := range files {
		err = g.parseStructFromFile(f)
		if err != nil {
			return err
		}
	}
	if g.MessageStruct == nil {
		return fmt.Errorf("не найдена структура event.message")
	}

	if len(g.GenStruct) == 0 {
		return fmt.Errorf("не найдено ни одной структуры для генерации схем")
	}

	for _, curStruct := range g.GenStruct {
		g.generateSchema(curStruct)
	}

	return nil
}

// parseStructFromFile - парсинг всех структур в файле.
func (g *Gen) parseStructFromFile(source *SourceFile) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, g.SourceDir+"/"+source.SourcePath(), nil, parser.ParseComments)
	if err != nil {
		return err
	}

	pkgName := node.Name.Name

	for _, gd := range node.Decls {
		targetStruct, ok := parseStruct(gd, pkgName)
		if !ok {
			continue
		}

		if err = targetStruct.Parse(); err != nil {
			return err
		}

		if targetStruct.Tags == nil {
			g.RefStruct[fmt.Sprintf("%s.%s", targetStruct.Pkg, targetStruct.Name)] = targetStruct
		} else if targetStruct.Tags.IsEventMessage {
			g.MessageStruct = targetStruct
		} else if targetStruct.Tags.Event != "" {
			g.GenStruct = append(g.GenStruct, targetStruct)
		} else {
			g.RefStruct[fmt.Sprintf("%s.%s", targetStruct.Pkg, targetStruct.Name)] = targetStruct
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

	pkgName := node.Name.Name

	for _, gd := range node.Decls {
		targetStruct, ok := parseStruct(gd, pkgName)
		if !ok {
			continue
		}

		if targetStruct.Tags == nil || targetStruct.Tags.Event == "" {
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

func (g *Gen) generateSchema(ss *schemaStruct) (*schema.Schema, bool) {
	res := schema.NewSchema()
	res.Title = ss.Tags.Event
	res.Description = ss.Tags.Description

	for _, sf := range ss.Fields {
		if prop, ok := g.generateFields(sf); ok {
			res.Properties = append(res.Properties, prop)
		}
	}

	return res, true
}

func (g *Gen) generateFields(sf *schemaField) (*schema.Property, bool) {
	// Check for integers
	var prop *schema.Property
	if strings.Contains(sf.Type, "int") {
		prop = schema.NewIntegerProperty(sf.Name, sf.Required)
	} else {
		//Check for other types
		switch sf.Type {
		case "string":
			prop = schema.NewStringProperty(sf.Name, sf.Required)
		case "bool":
			prop = schema.NewBoolProperty(sf.Name, sf.Required)
		case "Time":
			prop = schema.NewTimeProperty(sf.Name, sf.Required)
		case "struct":
			prop = schema.NewObjectProperty(sf.Name, sf.Required)
			/*
				for _, f := range sf.Fields {
					if p, ok := parseField(f); ok {
						prop.AddProperty(p)
					}
				}
			*/
		case "array":
			return nil, false
			//prop = schema.NewArrayProperty(sf.Name, sf.Required)
		default:
			log.Panicf("Field type %s not supported", sf.Type)
		}
	}
	return prop, true
}
