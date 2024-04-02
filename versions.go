package main

import (
	"fmt"
	"reflect"
)

var configVersions = map[int]interface{}{
	1: ConfigV1{},
}

var migrationUp = map[string]func(*ConfigV1) interface{}{
	"Version": func(c *ConfigV1) interface{} {
		return 2
	},
	"City": func(c *ConfigV1) interface{} {
		fmt.Println("QUI")
		return c.City + " " + c.Street.Name
	},
}

var migrationDown = map[string]func(*ConfigV2) interface{}{
	"Version": func(c *ConfigV2) interface{} {
		return 2
	},
	"City": func(c *ConfigV2) interface{} {
		fmt.Println("QUI")
		return c.City + " " + c.Street.Name
	},
}

// Address represents the address details
type Street struct {
	Field1 int    `yaml:"Field1"`
	Name   string `yaml:"Name" comment:"ciao prova indentazione\ntest bello" lineComment:"prova line"`
}

// Address represents the address details
type ConfigV1 struct {
	Version int    `yaml:"version"`
	Street  Street `yaml:"street" comment:"$comm1"`
	City    string `yaml:"city" comment:"Test comment for city" lineComment:"test"`
	ZipCode int    `yaml:"zipcode" comment:"Test comment for zipcode"`
}

type ConfigV2 struct {
	Version int    `yaml:"version"`
	Street  Street `yaml:"street" comment:"$comm1"`
	City    string `yaml:"city" comment:"Test comment for city" lineComment:"test"`
}

func up(source *ConfigV1, destination *ConfigV2, migration map[string]func(*ConfigV1) interface{}) {

	// Get the type and value of the destination struct
	destType := reflect.TypeOf(*destination)

	// Get the type and value of the source struct
	sourceType := reflect.TypeOf(*source)
	sourceValue := reflect.ValueOf(*source)

	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)

		if f, ok := migration[field.Name]; ok {
			modifyField(destination, field.Name, f(source))
		} else {
			if _, found := sourceType.FieldByName(field.Name); !found {
				panic("error while applying migration")
			}
			modifyField(destination, field.Name, sourceValue.FieldByName(field.Name).Interface())
		}
	}
}

func modifyField(obj interface{}, fieldName string, newValue interface{}) {
	value := reflect.ValueOf(obj).Elem() // Get the value of the struct
	fieldValue := value.FieldByName(fieldName)

	if !fieldValue.IsValid() {
		panic(fmt.Sprintf("Field %s not found\n", fieldName))
	}

	if !fieldValue.CanSet() {
		panic(fmt.Sprintf("Cannot set field %s value\n", fieldName))
	}

	newValueKind := reflect.ValueOf(newValue).Kind()

	if fieldValue.Kind() != newValueKind {
		panic(fmt.Sprintf("Type mismatch for field %s. Expected %s, got %s\n", fieldName, fieldValue.Kind(), newValueKind))
	}

	fieldValue.Set(reflect.ValueOf(newValue).Convert(fieldValue.Type()))
}
