// go build gogen/* && ./schemagen.exe pack/packer.go  pack/marshaller.go
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type tpl struct {
	FieldName string
}

var (
	funcGetSchema = template.Must(
		template.New("funcGetResponse").ParseFiles("../schema/message/v1.json"),
	)
)

func main() {
	/*
		schema, err := readCommonSchema("../schema/message/v1.json")
		if err != nil {
			log.Fatal(err)
		}
	*/

	//log.Println(schema)

	files, err := readFiles("./message/")

	for _, f := range files {
		err = genFromFile(f)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func readCommonSchema(file string) (map[string]any, error) {

	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	v := map[string]any{}
	err = json.Unmarshal(content, &v)

	return v, err
}

func genFromFile(sf *sourceFile) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "./message/"+sf.sourcePath(), nil, parser.ParseComments)
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
			fmt.Printf("SKIP %s GenDecl.Doc is nil\n", sf.sourcePath())
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
			targetStruct.generateSchema()
		}
	}

	fmt.Println(node.Name.Name)
	//funcGetResponse.Execute(out, tpl{node.Name.Name})

	return nil
}

func generateSchema(s *schemaStruct, f *sourceFile, baseSchema map[string]any) {
	out, err := os.Create("../schema/" + f.destPath())
	if err != nil {
		log.Fatal(err.Error())
	}
	defer out.Close()

	for _, spec := range s.Decl.Specs {
		fmt.Fprintln(out, "")
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

	def := baseSchema["definitions"].(map[string]any)
	data := def["event_data"].(map[string]any)
	for _, v := range s.Fields {
		data[v.Name] = v.ToMap()
	}

	j, err := json.Marshal(baseSchema)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Fprintln(out, string(j)) // empty line
}
