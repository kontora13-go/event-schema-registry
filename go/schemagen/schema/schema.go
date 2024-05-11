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
	Schema      string             `json:"$schema"`
	Id          string             `json:"$id,omitempty"`
	Title       string             `json:"title"`
	Description string             `json:"description,omitempty"`
	Type        Type               `json:"type"`
	Definitions map[string]*Schema `json:"definitions"`
	Properties  []*Property        `json:"properties"`
	Required    []string           `json:"required"`
}

func NewSchema() *Schema {
	return &Schema{
		Schema:     DefaultSchema,
		Type:       TypeObject,
		Properties: make([]*Property, 0),
		Required:   make([]string, 0),
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

func (s *Schema) AddProperty(p *Property) {
	s.Properties = append(s.Properties, p)
	if p.Required {
		s.Required = append(s.Required, p.Name)
	}
}
