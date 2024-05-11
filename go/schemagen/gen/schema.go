package gen

import (
	"fmt"
	"github.com/kontora13-go/event-schema-registry/schemagen/schema"
	"go/ast"
	"log"
	"reflect"
	"strings"
)

var CommentPrefix = "schemagen:"

type schemaStruct struct {
	Pkg    string
	Name   string
	Decl   *ast.GenDecl
	Schema *schema.Schema
	Fields []*schemaField
	Tags   *schemaTags
}

type schemaField struct {
	Name     string
	Type     string
	Required bool
	Ref      string
	Fields   []*schemaField
}

type schemaTags struct {
	Event          string
	Description    string
	IsEventMessage bool
	IsRef          bool
}

func newSchemaStruct() *schemaStruct {
	return &schemaStruct{
		Schema: schema.NewSchema(),
	}
}

func newSchemaField() *schemaField {
	return &schemaField{}
}

func newSchemaTags() *schemaTags {
	return &schemaTags{}
}

func (s *schemaStruct) Parse() error {
	//var err error

	/*
		for _, spec := range s.Decl.Specs {
			currType, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			currStruct, ok := currType.Type.(*ast.StructType)
			if !ok {
				continue
			}

			for _, field := range currStruct.Fields.List {
				if p, ok := parseField(field); ok {
					s.Schema.AddProperty(p)
				}
			}
		}

	*/

	return nil
}

func parseStruct(gd ast.Decl, pkgName string) (*schemaStruct, bool) {
	genD, ok := gd.(*ast.GenDecl)
	if !ok {
		fmt.Printf("SKIP %T is not *ast.GenDecl\n", gd)
		return nil, false
	}

	if genD.Doc == nil {
		//fmt.Printf("SKIP %s GenDecl.Doc is nil\n", source.SourcePath())
		return nil, false
	}

	if genD.Doc == nil {
		return nil, false
	}

	var ss *schemaStruct
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

		ss = newSchemaStruct()
		ss.Pkg = pkgName
		ss.Name = currType.Name.Name
		ss.Decl = genD

		ss.Fields = parseFields(currStruct.Fields)
	}

	if ss == nil {
		return nil, false
	}

	ss.Tags = parseTags(genD.Doc.List)

	return ss, true
}

func parseFields(fieldList *ast.FieldList) []*schemaField {
	fields := make([]*schemaField, 0)
	if fieldList.NumFields() == 0 {

	}
	for _, field := range fieldList.List {
		if f, ok := parseField(field); ok {
			fields = append(fields, f)
		}
	}

	return fields
}

func parseField(field *ast.Field) (*schemaField, bool) {
	if len(field.Names) == 0 {
		return nil, false
	}
	success := true

	sf := newSchemaField()
	if field.Tag != nil {
		// FIXME: первый параметр - всегда имя поля, остальные - теги
		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
		tagVal := tag.Get("schema")
		fmt.Println("schema:", tagVal)
		tagParams := strings.Split(tagVal, ",")

		for _, param := range tagParams {
			switch param {
			case "required":
				sf.Required = true
			case "not_null":
				//tableCol.NotNull = true
			case "-":
				success = false
				break
			default:
				sf.Name = param
			}
		}
	}

	if !success || sf.Name == "" {
		return nil, false
	}

	// Determine field type
	switch t := field.Type.(type) {
	case *ast.Ident:
		sf.Type = t.Name
	case *ast.SelectorExpr:
		sf.Type = t.Sel.Name
	case *ast.StructType:
		sf.Type = "struct"
		sf.Fields = parseFields(t.Fields)
	case *ast.ArrayType:
		sf.Type = "array"
	case *ast.StarExpr:
		x := t.X.(*ast.Ident)
		sf.Type = "ref"
		sf.Ref, _ = strings.CutPrefix(x.Name, "*")
	default:
		sf.Type = "unknown"
	}

	return sf, true
}

// parseTags - парсинг тегов схемы из комментариев к структуре
func parseTags(comments []*ast.Comment) *schemaTags {
	tag := newSchemaTags()
	hasTags := false
	for _, comment := range comments {
		currComment := strings.TrimSpace(strings.Replace(comment.Text, "//", "", 1))

		// Ищем комментарий с описанием схемы
		hasTags = hasTags || strings.HasPrefix(currComment, CommentPrefix)
		if !hasTags {
			continue
		}
		currComment = strings.TrimSpace(strings.Replace(currComment, CommentPrefix, "", 1))

		// Убираем префикс пары тега
		if !strings.HasPrefix(currComment, "#") {
			continue
		}
		currComment = strings.TrimSpace(strings.Replace(currComment, "#", "", 1))

		tagKeyValue := strings.Split(currComment, ":")
		if len(tagKeyValue) > 2 {
			log.Printf("некорректное описание тега: %v", comment.Text)
			continue
		}

		switch strings.TrimSpace(tagKeyValue[0]) {
		case "event":
			tag.Event = strings.TrimSpace(tagKeyValue[1])
		case "description":
			tag.Description = strings.TrimSpace(tagKeyValue[1])
		case "is_event_message":
			tag.IsEventMessage = true
		case "is_ref":
			tag.IsRef = true
		}
	}

	if !hasTags {
		return nil
	}

	return tag
}
