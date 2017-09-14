package serializer

import (
	"errors"
	"reflect"
	"strings"

	"google.golang.org/appengine/datastore"
)

var (
	ErrBadSchemaStructTag = errors.New("serializer: Bad serializer struct tag format")
	ErrInvalidEntityType  = errors.New("serializer: invalid entity type")
	ErrNoMoreProperties   = errors.New("serializer: No more properties found")
)

const (
	annotationSchema   = "serializer"
	annotationPrimary  = "primary"
	annotationRelation = "relation"
)

// LoadEntity loads entity from property list
func LoadEntity(entity interface{}, ps datastore.PropertyList) error {
	value := reflect.ValueOf(entity)
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return ErrInvalidEntityType
	}

	// Copy property list
	properties := ps

	return loadEntity(entity, &properties, "")
}

// SaveEntity returns a PropertyList from an entity
// The recursion level - Level of nested entity - is limited to 2
func SaveEntity(entity interface{}) (datastore.PropertyList, error) {
	value := reflect.ValueOf(entity)
	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		return nil, ErrInvalidEntityType
	}

	return saveEntity(entity, "", false, 0)
}

func loadEntity(entity interface{}, ps *datastore.PropertyList, namespace string) error {
	// Load entity properties
	key, properties := extractEntityProperties(namespace, ps)

	if len(properties) == 0 {
		return ErrNoMoreProperties
	}

	for i := 0; i < len(properties); i++ {
		property := &properties[i]
		property.Multiple = false
	}

	// Load entity fields
	err := datastore.LoadStruct(entity, properties)
	if err != nil {
		return err
	}

	entityValue := reflect.ValueOf(entity).Elem()

	// Set ID if available
	if key != nil {
		entityValue.FieldByName("ID").Set(reflect.ValueOf(key.Encode()))
	}

	// Look for relations
	for i := 0; i < entityValue.NumField(); i++ {
		fieldValue := entityValue.Field(i)
		fieldType := entityValue.Type().Field(i)
		tag := fieldType.Tag.Get(annotationSchema)

		if tag == "" {
			continue
		}

		args := strings.Split(tag, ",")
		if len(args) < 2 {
			return ErrBadSchemaStructTag
		}
		annotation := args[0]

		if annotation == annotationRelation {
			currentNamespace := strings.Trim(namespace+"."+args[1], ".")
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				value := reflect.New(fieldValue.Type().Elem())
				valueI := value.Interface()
				err = loadEntity(valueI, ps, currentNamespace)
				if err != nil && err != ErrNoMoreProperties {
					return err
				} else if err == nil {
					fieldValue.Set(value)
				}
			} else if fieldValue.Kind() == reflect.Slice {
				for {
					value := reflect.New(reflect.TypeOf(fieldValue.Interface()).Elem().Elem())
					valueI := value.Interface()
					err = loadEntity(valueI, ps, currentNamespace)
					if err != nil {
						if err == ErrNoMoreProperties {
							break
						} else {
							return err
						}
					}
					fieldValue.Set(reflect.Append(fieldValue, value))
				}
			}
		}
	}
	return nil
}

func saveEntity(entity interface{}, namespace string, multiple bool, level int) (datastore.PropertyList, error) {

	if level > 2 {
		return datastore.PropertyList{}, nil
	}

	entityValue := reflect.ValueOf(entity).Elem()

	// Save default entity attributes
	ps, err := datastore.SaveStruct(entity)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ps); i++ {
		property := &ps[i]
		property.Name = strings.Trim(namespace+"."+property.Name, ".")
		property.Multiple = multiple
	}

	for i := 0; i < entityValue.NumField(); i++ {
		fieldValue := entityValue.Field(i)
		fieldType := entityValue.Type().Field(i)
		tag := fieldType.Tag.Get(annotationSchema)

		if tag == "" {
			continue
		}

		args := strings.Split(tag, ",")
		if len(args) < 2 {
			return nil, ErrBadSchemaStructTag
		}

		annotation := args[0]

		if annotation == annotationPrimary {
			id := fieldValue.Interface().(string)
			if id != "" && namespace != "" {
				key, err := datastore.DecodeKey(fieldValue.Interface().(string))
				if err != nil {
					return nil, err
				}
				idProperty := datastore.Property{
					Name:     strings.Trim(namespace+".id", "."),
					Value:    key,
					Multiple: multiple,
				}
				ps = append(ps, idProperty)
			}
		} else if annotation == annotationRelation {
			currentNamespace := strings.Trim(namespace+"."+args[1], ".")
			if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() {
				properties, err := saveEntity(fieldValue.Interface(), currentNamespace, multiple == true, level+1)
				if err != nil {
					return nil, err
				}
				ps = append(ps, properties...)
			} else if fieldValue.Kind() == reflect.Slice {
				for i := 0; i < fieldValue.Len(); i++ {
					properties, err := saveEntity(fieldValue.Index(i).Interface(), currentNamespace, true, level+1)
					if err != nil {
						return nil, err
					}
					ps = append(ps, properties...)
				}
			}
		}
	}
	return ps, nil
}

func extractEntityProperties(namespace string, ps *datastore.PropertyList) (*datastore.Key, datastore.PropertyList) {
	properties := datastore.PropertyList{}
	var key *datastore.Key
	propertyMap := map[string]bool{}

	deleted := 0
	for i := range *ps {
		j := i - deleted
		property := (*ps)[j]
		index := strings.LastIndex(property.Name, ".")
		hasNamespace := index != -1
		currentNamespace := ""
		if hasNamespace {
			currentNamespace = property.Name[:index]
		}
		if currentNamespace == namespace {
			name := strings.TrimPrefix(property.Name, currentNamespace+".")
			if propertyMap[name] == false {
				if name == "id" {
					if key == nil {
						key = property.Value.(*datastore.Key)
					}
				} else {
					property.Name = name
					properties = append(properties, property)
				}
				propertyMap[name] = true
				*ps = (*ps)[:j+copy((*ps)[j:], (*ps)[j+1:])]
				deleted++
			}
		}
	}
	return key, properties
}
