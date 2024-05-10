package main

import (
	"encoding/json"
	"fmt"
	"github.com/kontora13-go/event-schema-registry/schemagen/schema"
	"go/ast"
	"log"
	"os"
	"reflect"
	"strings"
)

type schemaStruct struct {
	Name   string
	Path   string
	Decl   *ast.GenDecl
	Schema *schema.Schema
	Fields []*schemaField
}

type schemaField struct {
	Name     string
	Type     schemaFieldType
	Required bool
}

func (s *schemaField) ToMap() map[string]any {
	m := map[string]any{
		"type": s.Type,
	}

	return m
}

type schemaFieldType string

const (
	_TypeNumber = "number"
	_TypeString = "string"
	_TypeBool   = "bool"
	_TypeTime   = "time"
	_TypeRef    = "ref"
)

type schemaObject struct {
	Properties struct {
		EventId struct {
			Type string `json:"type"`
		} `json:"event_id"`
		EventName struct {
			Type string `json:"type"`
		} `json:"event_name"`
		EventTime struct {
			Type string `json:"type"`
		} `json:"event_time"`
		EventVersion struct {
			Enum []int `json:"enum"`
		} `json:"event_version"`
		Producer struct {
			Type string `json:"type"`
		} `json:"producer"`
		TraceId struct {
			Type string `json:"type"`
		} `json:"trace_id"`
	} `json:"properties"`
	Required []string `json:"required"`
	Type     string   `json:"type"`
}

func (s *schemaStruct) generateSchema() {
	out, err := os.Create("../schema/" + s.Path)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()

	for _, spec := range s.Decl.Specs {
		_, err = fmt.Fprintln(out, "")
		if err != nil {
			log.Fatal(err)
		}

		currType, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		currStruct, ok := currType.Type.(*ast.StructType)
		if !ok {
			continue
		}

		for _, field := range currStruct.Fields.List {
			if len(field.Names) == 0 {
				continue
			}

			sField := &schemaField{}
			if field.Tag != nil {
				preventThisField := false
				tag := reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
				tagVal := tag.Get("json")
				fmt.Println("dbe:", tagVal)
				tagParams := strings.Split(tagVal, ",")
			PARAMSLOOP:
				for _, param := range tagParams {
					switch param {
					case "required":
						sField.Required = true
					case "not_null":
						//tableCol.NotNull = true
					case "-":
						preventThisField = true
						break PARAMSLOOP
					default:
						sField.Name = param
					}

				}
				if preventThisField {
					continue
				}
			}

			if sField.Name == "" {
				continue
			}

			/*
				if tableCol.ColName == "" {
					tableCol.ColName = tableCol.FieldName
				}
				if fieldIsPrimKey {
					curTable.PrimaryKey = tableCol
				}

			*/
			//Determine field type
			var fieldType string
			switch field.Type.(type) {
			case *ast.Ident:
				fieldType = field.Type.(*ast.Ident).Name
			case *ast.SelectorExpr:
				fieldType = field.Type.(*ast.SelectorExpr).Sel.Name
			}
			//fieldType := field.Type.(*ast.Ident).Name
			//fmt.Printf("%s- %s\n", tableCol.FieldName, fieldType)
			//Check for integers
			if strings.Contains(fieldType, "int") {
				sField.Type = _TypeNumber
			} else {
				//Check for other types
				switch fieldType {
				case "string":
					sField.Type = _TypeString
				case "bool":
					sField.Type = _TypeBool
				case "Time":
					sField.Type = _TypeTime
				case "Struct":
					sField.Type = _TypeRef
				case "Slice":
				//	tableCol.ColType = "TIMESTAMP"
				default:
					log.Panicf("Field type %s not supported", fieldType)
				}
			}
			//tableCol.FieldType = fieldType
			//curTable.Columns = append(curTable.Columns, tableCol)
			//curTable.StructName = currType.Name.Name
			s.Fields = append(s.Fields, sField)
		}
	}

	/*

		def := baseSchema["definitions"].(map[string]any)
		data := def["event_data"].(map[string]any)
		for _, v := range s.Fields {
			data[v.Name] = v.ToMap()
		}

	*/

	j, err := json.Marshal(s.Schema)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Fprintln(out, string(j)) // empty line
}
