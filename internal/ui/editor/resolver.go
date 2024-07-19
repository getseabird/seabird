package editor

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type resolver struct {
	*openapi3.T
}

func (r *resolver) typeName(schema *openapi3.SchemaRef) string {
	var typ string
	if schema.Value.Type != nil {
		typ = schema.Value.Type.Slice()[0]
	}

	if schema.Ref != "" {
		return refName(schema.Ref)
	}

	switch typ {
	case openapi3.TypeString, openapi3.TypeNumber, openapi3.TypeInteger, openapi3.TypeBoolean, openapi3.TypeNull:
		return typ
	case openapi3.TypeArray:
		return fmt.Sprintf("[]%s", r.typeName(schema.Value.Items))
	}

	if subtypes := r.subtypes(schema); len(subtypes) > 0 {
		var names []string
		for _, s := range subtypes {
			names = append(names, r.typeName(s))
		}
		return strings.Join(names, " | ")
	}

	return typ
}

func (r *resolver) subtypes(schema *openapi3.SchemaRef) (ret []*openapi3.SchemaRef) {
	types := []*openapi3.SchemaRef{}
	if schema.Value.Items != nil {
		types = append(types, schema.Value.Items)
	}
	types = append(types, schema.Value.OneOf...)
	types = append(types, schema.Value.AnyOf...)
	types = append(types, schema.Value.AllOf...)
	for _, typ := range types {
		types = append(types, r.subtypes(typ)...)
	}
	for _, typ := range types {
		if s := r.resolve(typ.Ref); s != nil {
			ret = append(ret, s)
		} else if typ.Value != nil {
			ret = append(ret, typ)
		}
	}
	return
}

func (r *resolver) resolve(ref string) *openapi3.SchemaRef {
	schema := r.Components.Schemas[strings.TrimPrefix(ref, "#/components/schemas/")]
	if schema != nil {
		schema.Ref = ref
	}
	return schema
}
