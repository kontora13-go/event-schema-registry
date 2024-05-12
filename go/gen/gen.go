package gen

import (
	"encoding/json"
	"fmt"
	schema2 "github.com/kontora13-go/event-schema-registry/schema"
	"go/parser"
	"go/token"
	"log"
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

	if err = cleanFiles(g.DestDir); err != nil {
		return fmt.Errorf("ошибка удаления старых схем в '%s': %s", g.DestDir, err.Error())
	}

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
		err = g.saveMessageSchema(curStruct)
		if err != nil {
			return fmt.Errorf("не удалось сгенерировать схему: %v", err.Error())
		}
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

		if targetStruct.Tags != nil {
			if targetStruct.Tags.IsEventMessage {
				g.MessageStruct = targetStruct
			} else if targetStruct.Tags.Event != "" {
				targetStruct.File = source
				g.GenStruct = append(g.GenStruct, targetStruct)
			}
		}
		g.RefStruct[fmt.Sprintf("%s.%s", targetStruct.Pkg, targetStruct.Name)] = targetStruct
	}

	return nil
}

// generateSchema - генерация схемы по результатам парсинга
func (g *Gen) generateSchema(ss *schemaStruct) (*schema2.Schema, bool) {
	if ss.Schema != nil {
		return ss.Schema, true
	}

	ss.Schema = schema2.NewSchema()
	//ss.Schema.Id = ""
	if ss.Tags != nil {
		ss.Schema.Title = ss.Tags.Event
		ss.Schema.Description = ss.Tags.Description
	}

	for _, sf := range ss.Fields {
		if prop, ok := g.generateProperty(ss, sf); ok {
			ss.Schema.AddProperty(prop)
		}
	}

	return ss.Schema, true
}

// generateProperty - генерация поля схемы по результатам парсинга
func (g *Gen) generateProperty(ss *schemaStruct, sf *schemaField) (*schema2.Property, bool) {
	// Check for integers
	var prop *schema2.Property
	if strings.Contains(sf.Type, "int") {
		prop = schema2.NewIntegerProperty(sf.Name, sf.Required)
	} else {
		//Check for other types
		switch sf.Type {
		case "string":
			prop = schema2.NewStringProperty(sf.Name, sf.Required)
		case "bool":
			prop = schema2.NewBoolProperty(sf.Name, sf.Required)
		case "Time":
			prop = schema2.NewTimeProperty(sf.Name, sf.Required)
		case "struct":
			prop = schema2.NewObjectProperty(sf.Name, sf.Required)
			for _, f := range sf.Fields {
				if p, ok := g.generateProperty(ss, f); ok {
					prop.AddProperty(p)
				}
			}
		case "array":
			return nil, false
			//prop = schema.NewArrayProperty(sf.Name, sf.Required)
		case "ref":
			prop = schema2.NewRefProperty(sf.Name, sf.Required)
			prop.RefId = fmt.Sprintf("#/definitions/%s", sf.Name)
			var ok bool
			if prop.Ref, ok = g.findRefSchema(ss.Pkg, sf.Ref); !ok {
				log.Panicf("не удалось сгенерировать схему для типа %v", sf.Ref)
			}
		default:
			log.Panicf("Field type %s not supported", sf.Type)
		}
	}
	return prop, true
}

// findRefSchema - поиск ссылки на схему
func (g *Gen) findRefSchema(pkg string, ref string) (*schema2.Schema, bool) {
	if strings.Index(ref, ".") < 0 {
		ref = fmt.Sprintf("%s.%s", pkg, ref)
	}
	ss, ok := g.RefStruct[ref]
	if !ok {
		return nil, false
	}

	return g.generateSchema(ss)
}

// saveMessageSchema - подготовка и сохранение схемы Event-message
func (g *Gen) saveMessageSchema(ss *schemaStruct) error {
	if ss.Tags == nil {
		return fmt.Errorf("не определены параметры генерации 'genschema:' для %s.%s", ss.Pkg, ss.Name)
	}

	eventSchema := *g.MessageStruct
	for _, v := range eventSchema.Fields {
		if v.EventData {
			v.Ref = fmt.Sprintf("%s.%s", ss.Pkg, ss.Name)
			v.Type = schema2.TypeRef
			break
		}
	}

	res, ok := g.generateSchema(&eventSchema)
	if !ok {
		return fmt.Errorf("не удалось сгенерировать схему")
	}

	res.Schema = schema2.DefaultSchema
	res.Title = ss.Tags.Event
	res.Description = ss.Tags.Description

	for _, v := range res.Properties {
		if v.Type != schema2.TypeRef {
			continue
		}

		res.Definitions[v.Name] = v.Ref
	}

	path := make([]string, 0)
	path = append(path, g.DestDir)
	for _, v := range ss.File.Path {
		path = append(path, v)
	}
	e := strings.Split(ss.Tags.Event, ".")
	file := e[len(e)-1] + ".json"
	for i := 0; i < len(e)-1; i++ {
		path = append(path, e[i])
	}

	j, err := json.MarshalIndent(&res, "", "  ")
	if err != nil {
		log.Fatal(err.Error())
	}
	err = saveFile(strings.Join(path, "/"), file, j)

	return err
}
