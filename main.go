package main

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

var longComments = map[string]string{
	"$comm1": "Test commento numeor 1",
}

type Street struct {
	Field1 int    `yaml:"Field1"`
	Name   string `yaml:"Name" comment:"ciao prova indentazione\ntest bello" lineComment:"prova line"`
}

// Address represents the address details
type Address struct {
	Version int    `yaml:"version"`
	Street  Street `yaml:"street" comment:"$comm1"`
	City    string `yaml:"city" comment:"Test comment for city" lineComment:"test"`
	ZipCode int    `yaml:"zipcode" comment:"Test comment for zipcode"`
}

func main() {
	// Create an example Address object with test values
	address := Address{
		Version: 1,
		Street: Street{
			Field1: 444,
			Name:   "iohfnocoijas",
		},
		City:    "Testville",
		ZipCode: 4421243,
	}
	WriteYaml(address, "address.yaml")
	config, version := LoadConfigVersioned("address.yaml")
	fmt.Printf("config version %d: %#v", version, config)
	if _, ok := config.(Address); ok {
		fmt.Println("Is of correct Type")
	} else {
		panic("NO")
	}
}

// WriteYaml create a yaml file with name [name] from the tagged [data] struct
func WriteYaml(data interface{}, path string) {
	// Create a YAML nodes representation of the Address struct
	yamlObject, err := GenerateYAMLobject(data)
	if err != nil {
		panic(fmt.Errorf("error generating YAML: %w", err))
	}
	yamlBytes, err := yaml.Marshal(yamlObject)
	if err != nil {
		panic(fmt.Errorf("error marshalling YAML: %w", err))
	}

	// Write the YAML to a file
	file, err := os.Create(path)
	if err != nil {
		panic(fmt.Errorf("error creating file: %w", err))
	}
	defer file.Close()

	_, err = file.Write(yamlBytes)
	if err != nil {
		panic(fmt.Errorf("error writing YAML to file: %w", err))
	}
	fmt.Printf("%v file created successfully.", path)
}

// generateYAMLobject generates Node object formatted for a yaml file
func GenerateYAMLobject(data interface{}) (*yaml.Node, error) {
	// Get the type of the data
	dataType := reflect.TypeOf(data)

	// Create a new YAML node for the root
	rootNode := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	nextIndent := false
	// Iterate over the fields of the data struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		commentTag := field.Tag.Get("comment")         // Get the comment tag value
		lineCommentTag := field.Tag.Get("lineComment") // Get the lineComment tag value

		if com, ok := longComments[commentTag]; ok { // Check if comment is key of a long comment and subsitute it
			commentTag = com
		}

		if com, ok := longComments[lineCommentTag]; ok { // Check if lineComment is key of a long comment and subsitute it
			lineCommentTag = com
		}

		fieldName := field.Tag.Get("yaml") // Get the yaml tag value
		if fieldName == "" {
			fieldName = field.Name // Use the field name as the key if yaml tag is empty
		}

		if nextIndent {
			commentTag = "\n" + commentTag // Put new linebe before start of new struct
			nextIndent = false
		}

		// Create key node
		keyNode := &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fieldName,
			HeadComment: commentTag,
		}

		var valueNode *yaml.Node
		var err error
		// If field is of type struct
		if reflect.ValueOf(data).Field(i).Type().Kind() == reflect.Struct {
			valueNode, err = GenerateYAMLobject(reflect.ValueOf(data).Field(i).Interface())
		} else if field.Type.Kind() == reflect.Ptr { // If field is of type pointer
			valueNode, err = GenerateYAMLobject(reflect.ValueOf(data).Field(i).Elem().Interface())
			keyNode.HeadComment = "\n" + keyNode.HeadComment
			nextIndent = true
		} else {
			// Create value node
			valueNode = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       fmt.Sprintf("%v", reflect.ValueOf(data).Field(i).Interface()), // Get the field value from the struct
				LineComment: lineCommentTag,
			}
		}
		if err != nil {
			return nil, err
		}

		// Append key and value nodes to the root node
		rootNode.Content = append(rootNode.Content, keyNode, valueNode)
	}

	return rootNode, nil
}

// LoadYAML loads a yaml file as object
func LoadYAML(path string, object interface{}) {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("error opening file: %w", err))
	}
	defer file.Close()

	// Read the content of the file
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(object); err != nil {
		panic(fmt.Errorf("error decoding YAML: %w", err))
	}
}

// GetVersion returns version of config file
func GetVersion(path string) int {
	var data interface{}
	LoadYAML(path, &data)

	// Assert data to a map[string]interface{}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		panic("invalid YAML format")
	}

	// Handle different versions based on the "version" field
	version, ok := dataMap["version"].(int)
	if !ok {
		panic("version field is not an integer")
	}
	return version
}

// LoadConfigVersioned loads a config file as a struct of the related version and returns also the version
func LoadConfigVersioned(path string) (interface{}, int) {
	version := GetVersion(path)

	// Get the appropriate struct type based on version
	configType, ok := configVersions[version]
	if !ok {
		panic(fmt.Sprintf("unsupported version: %d", version))
	}

	// Load YAML into the appropriate struct type
	configValue := reflect.New(reflect.TypeOf(configType)).Interface()
	LoadYAML(path, configValue)
	config := reflect.Indirect(reflect.ValueOf(configValue)).Interface()

	return config, version
}
