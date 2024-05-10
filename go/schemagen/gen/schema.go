package gen

import (
	"fmt"
	"github.com/kontora13-go/event-schema-registry/schemagen/schema"
	"go/ast"
	"log"
	"reflect"
	"strings"
)

type schemaStruct struct {
	Name   string
	Decl   *ast.GenDecl
	Schema *schema.Schema
}

type schemaField struct {
	Name     string
	Type     string
	Required bool
	Fields   []*ast.Field
}

func newSchemaStruct() *schemaStruct {
	return &schemaStruct{
		Schema: schema.NewSchema(),
	}
}

func (s *schemaStruct) Parse() error {
	//var err error

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

	return nil
}

func parseField(field *ast.Field) (*schema.Property, bool) {
	if len(field.Names) == 0 {
		return nil, false
	}
	success := true

	sf := &schemaField{}
	if field.Tag != nil {
		tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
		tagVal := tag.Get("schema")
		fmt.Println("schema:", tagVal)
		tagParams := strings.Split(tagVal, ",")

	PARAMSLOOP:
		for _, param := range tagParams {
			switch param {
			case "required":
				sf.Required = true
			case "not_null":
				//tableCol.NotNull = true
			case "-":
				success = false
				break PARAMSLOOP
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
		sf.Fields = t.Fields.List
	case *ast.ArrayType:
		sf.Type = "array"
	}

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
			for _, f := range sf.Fields {
				if p, ok := parseField(f); ok {
					prop.AddProperty(p)
				}
			}
		case "array":
			return nil, false
			//prop = schema.NewArrayProperty(sf.Name, sf.Required)
		default:
			log.Panicf("Field type %s not supported", sf.Type)
		}
	}
	return prop, true
}
