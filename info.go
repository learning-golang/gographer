package gographer

import (
	"reflect"
	"github.com/graphql-go/graphql"
	"fmt"
	"github.com/graphql-go/relay"
)

type SchemaInfo struct {
	types        []*TypeInfo
	typesByName  map[string]*TypeInfo
	rootInstance interface{}
	mutationInstance interface{}
}

func NewSchemaInfo() *SchemaInfo {
	return &SchemaInfo{
		typesByName: make(map[string]*TypeInfo),
	}
}

func (sch *SchemaInfo) RegType(instance interface{}) *TypeInfo {
	typeDef := NewTypeInfo(instance)
	sch.types = append(sch.types, typeDef)
	sch.typesByName[typeDef.Name] = typeDef
	return typeDef
}

type TypeInfo struct {
	Name           string
	Type           reflect.Type
	idResolver     IDResolver
	fields         graphql.Fields
	resolvedFields []ResolvedFieldInfo
	mutationFields []MutationFieldInfo
	isRootType     bool
	isMutationType bool
	instance       interface{}
}

type IDResolver func(id string) interface{}

func NewTypeInfo(instance interface{}) *TypeInfo {
	type_ := reflect.TypeOf(instance)
	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	typeDef := TypeInfo{
		Type: type_,
		Name: type_.Name(),
		fields: make(graphql.Fields),
		instance: instance,
	}
	return &typeDef
}

func (typ *TypeInfo) SetIDResolver(f IDResolver) *TypeInfo {
	typ.idResolver = f
	return typ
}

func (typ *TypeInfo) SetRoot() *TypeInfo {
	typ.isRootType = true
	return typ
}

func (typ *TypeInfo) SimpleField(name string) *TypeInfo {
	for i := 0; i < typ.Type.NumField(); i++ {
		field := typ.Type.Field(i)
		if field.Name == name || field.Tag.Get("json") == name {
			if qlType := ToQLType(field.Type); qlType != nil {
				return typ.AddField(name, &graphql.Field{
					Type: qlType,
				})
			}
		}
	}
	Warning("SimpleField not found", typ.Name, name)
	return typ
}

func (typ *TypeInfo) IDField(name string, idFetcher relay.GlobalIDFetcherFn) *TypeInfo {
	for i := 0; i < typ.Type.NumField(); i++ {
		field := typ.Type.Field(i)
		if field.Name == name || field.Tag.Get("json") == name {
			return typ.AddField(name, relay.GlobalIDField(typ.Name, idFetcher))
		}
	}
	Warning("IDField not found", typ.Name, name)
	return typ
}

func (typ *TypeInfo) ResolvedField(name string, methodName string, args ...ArgInfo) *TypeInfo {
	typ.resolvedFields = append(typ.resolvedFields, ResolvedFieldInfo{
		Name: name,
		MethodName: methodName,
		Args: args,
	})
	return typ
}

type ArgInfo struct {
	Name         string
	DefaultValue interface{}
	NonNull      bool
}

func (typ *TypeInfo) AddField(name string, field *graphql.Field) *TypeInfo {
	typ.fields[name] = field
	return typ
}

type ResolvedFieldInfo struct {
	Name       string
	MethodName string
	Args       []ArgInfo
}

func (typ *TypeInfo) SetMutation() *TypeInfo {
	typ.isMutationType = true
	return typ
}

func (typ *TypeInfo) MutationField(name string, methodName string, args []ArgInfo, outputs []OutputInfo) *TypeInfo {
	typ.mutationFields = append(typ.mutationFields, MutationFieldInfo{
		Name: name,
		MethodName: methodName,
		Args: args,
		Outputs: outputs,
	})
	return typ
}

type MutationFieldInfo struct {
	Name       string
	MethodName string
	Args       []ArgInfo
	Outputs    []OutputInfo
}

type OutputInfo struct {
	Name          string
	ElemInterface interface{}
}

func ToQLType(typ reflect.Type) graphql.Output {
	switch typ.Kind() {
	case reflect.Float32:fallthrough
	case reflect.Float64:
		return graphql.Float
	case reflect.String:
		return graphql.String
	case reflect.Bool:
		return graphql.Boolean
	case reflect.Int:fallthrough
	case reflect.Int8:fallthrough
	case reflect.Int16:fallthrough
	case reflect.Int32:fallthrough
	case reflect.Int64:fallthrough
	case reflect.Uint:fallthrough
	case reflect.Uint8:fallthrough
	case reflect.Uint16:fallthrough
	case reflect.Uint32:fallthrough
	case reflect.Uint64:
		return graphql.Int
	default:
		return nil
	}
}

func Warning(a ...interface{}) {
	a = append([]interface{}{"[Gographer warning]"}, a...)
	fmt.Println(a...)
}
