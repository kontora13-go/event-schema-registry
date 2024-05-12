package schema

import "encoding/json"

const DefaultSchema = "https://json-schema.org/draft/2020-12/schema"

type Type string

const (
	TypeObject  = "object"
	TypeNumber  = "number"
	TypeInteger = "integer"
	TypeString  = "string"
	TypeBool    = "bool"
	TypeTime    = "time"
	TypeRef     = "ref"
)

type Schema struct {
	Schema      string               `json:"$schema,omitempty"`
	Id          string               `json:"$id,omitempty"`
	Title       string               `json:"title,omitempty"`
	Description string               `json:"description,omitempty"`
	Type        Type                 `json:"type"`
	Definitions map[string]*Schema   `json:"definitions,omitempty"`
	Properties  map[string]*Property `json:"properties"`
	Required    []string             `json:"required"`
}

func NewSchema() *Schema {
	return &Schema{
		Type:        TypeObject,
		Definitions: make(map[string]*Schema),
		Properties:  make(map[string]*Property),
		Required:    make([]string, 0),
	}
}

func (s *Schema) _MarshalJSON() ([]byte, error) {
	s.Required = make([]string, 0, len(s.Properties))
	for _, v := range s.Properties {
		if v == nil {
			continue
		}
		if v.Required {
			s.Required = append(s.Required, v.Name)
		}
	}

	return json.Marshal(s)
}

func (s *Schema) AddProperty(prop *Property) {
	s.Properties[prop.Name] = prop
	if prop.Required {
		s.Required = append(s.Required, prop.Name)
	}
}
