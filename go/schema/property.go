package schema

import "encoding/json"

type Property struct {
	Name        string               `json:"name,omitempty"`
	Description string               `json:"description,omitempty"`
	Type        Type                 `json:"type,omitempty"`
	Required    bool                 `json:"-"`
	RefId       string               `json:"$ref,omitempty"`
	Ref         *Schema              `json:"-"`
	Properties  map[string]*Property `json:"properties,omitempty"`
}

func (p *Property) MarshalJSON() ([]byte, error) {
	var m map[string]any
	switch p.Type {
	case TypeObject:
		m = p.mappingObject()
	case TypeString:
		m = p.mappingString()
	case TypeRef:
		m = p.mappingRef()
	default:
		m = p.mappingDefault()
	}

	return json.Marshal(m)
}

func (p *Property) mappingDefault() map[string]any {
	m := map[string]any{
		"type": p.Type,
	}
	if p.Description != "" {
		m["description"] = p.Description
	}

	return m
}

/*
 String
*/

func NewStringProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeString,
		Required: required,
	}
}

func (p *Property) mappingString() map[string]any {
	return p.mappingDefault()
}

/*
 Number, Integer
*/

func NewNumberProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeNumber,
		Required: required,
	}
}

func NewIntegerProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeInteger,
		Required: required,
	}
}

/*
 Time
*/

func NewTimeProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeTime,
		Required: required,
	}
}

/*
 Bool
*/

func NewBoolProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeBool,
		Required: required,
	}
}

/*
 Object
*/

func NewObjectProperty(name string, required bool) *Property {
	return &Property{
		Name:       name,
		Type:       TypeObject,
		Required:   required,
		Properties: make(map[string]*Property),
	}
}

func (p *Property) mappingObject() map[string]any {
	m := p.mappingDefault()
	prop := make(map[string]any, len(p.Properties))
	req := make([]string, 0, len(p.Properties))
	for _, v := range p.Properties {
		prop[v.Name] = v
		if v.Required {
			req = append(req, v.Name)
		}
	}
	m["properties"] = prop
	if len(req) > 0 {
		m["required"] = req
	}

	return m
}

func (p *Property) AddProperty(prop *Property) {
	p.Properties[prop.Name] = prop
}

/*
 Ref
*/

func NewRefProperty(name string, required bool) *Property {
	return &Property{
		Name:     name,
		Type:     TypeRef,
		Required: required,
	}
}

func (p *Property) mappingRef() map[string]any {
	m := map[string]any{
		"$ref": p.RefId,
	}
	if p.Description != "" {
		m["description"] = p.Description
	}

	return m
}
