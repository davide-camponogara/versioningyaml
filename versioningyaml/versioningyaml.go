package versioningyaml

import (
	"fmt"
	"os"
	"reflect"
	"versioningyaml/utils"

	"gopkg.in/yaml.v3"
)

// ConfigVersions contains an ordered list of the versions of the yaml file
// every entry is a ConfigVersion object that contains:
//   - the Config struct
//   - the CustomMigration UP
//   - the CustomMigration DOWN
//
// the custom migrations are unes when one have to manipulate the fileds instead
// if a field doesn't change from the version before or is one is a new (standalone) field a CustomMigration is not needed
var configVersions []utils.ConfigVersion

// SetConfigVersions setter for ConfigVersions
//
// ConfigVersions contains an ordered list of the versions of the yaml file
// every entry is a ConfigVersion object that contains:
//   - the Config struct
//   - the CustomMigration UP
//   - the CustomMigration DOWN
//
// the custom migrations are unes when one have to manipulate the fileds instead
// if a field doesn't change from the version before or is one is a new (standalone) field a CustomMigration is not needed
func SetConfigVersions(cv []utils.ConfigVersion) {
	configVersions = cv
}

// LongComments is a map containing long comments
// by convenction a reference to a long comment is denoted with a $ in form of the name
var longComments map[string]string

// SetLongComments setter for LongComments
//
// LongComments is a map containing long comments
// by convenction a reference to a long comment is denoted with a $ in form of the name
func SetLongComments(lc map[string]string) {
	longComments = lc
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

// getVersion returns version of config file
func getVersion(path string) int {
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
	version := getVersion(path)

	// Get the appropriate struct type based on version
	configType, _ := findByVersion(version)
	if configType == nil {
		panic("error finding version")
	}

	// Load YAML into the appropriate struct type
	configValue := reflect.New(reflect.TypeOf(configType.Config)).Interface()
	LoadYAML(path, configValue)
	config := reflect.Indirect(reflect.ValueOf(configValue)).Interface()

	return config, version
}

// Migrate apply migration from config source to config destination objects
func MigrateOne(source interface{}, destination interface{}, migration utils.CustomMigration) {
	// Get the type and value of the destination struct
	destValue := reflect.ValueOf(destination).Elem()

	// Get the type and value of the source struct
	sourceValue := reflect.ValueOf(source).Elem()

	for i := 0; i < destValue.NumField(); i++ {
		field := destValue.Type().Field(i)
		fieldName := field.Name

		if fieldName == "Version" {
			destField := destValue.FieldByName(fieldName)
			if destField.IsValid() && destField.CanSet() {
				destField.Set(reflect.ValueOf(destination.(utils.Config).V()))
			}
		} else if f, ok := migration[fieldName]; ok {
			newValue := f(sourceValue.Interface())
			destField := destValue.FieldByName(fieldName)
			if destField.IsValid() && destField.CanSet() {
				destField.Set(reflect.ValueOf(newValue))
			}
		} else {
			sourceField := sourceValue.FieldByName(fieldName)
			if sourceField.IsValid() {
				destField := destValue.FieldByName(fieldName)
				if destField.IsValid() && destField.CanSet() {
					destField.Set(sourceField)
				}
			} else {
				panic(fmt.Sprintf("error while migrating value: %v", fieldName))
			}
		}
	}
}

func findByVersion(version int) (*utils.ConfigVersion, int) {
	for i, cv := range configVersions {
		if cv.Config.V() == version {
			return &cv, i
		}
	}
	return nil, -1
}

// MigrateUp applies the UP migrations form yaml [source] to [destination] returning the
// fullfilled [destination] version (Config interface)
func MigrateUp(source interface{}, destination interface{}) utils.Config {
	var vStart, vFinish int
	if s, ok := source.(utils.Config); ok {
		_, vStart = findByVersion(s.V())
	}
	if d, ok := destination.(utils.Config); ok {
		_, vFinish = findByVersion(d.V())
	}

	var current = source
	for i := vStart; i < vFinish; i++ {
		next := reflect.New(reflect.TypeOf(configVersions[i+1].Config)).Interface()
		MigrateOne(current, next, configVersions[i+1].Up)
		current = next
	}
	return current.(utils.Config)
}

// MigrateDown applies the DOWN migrations form yaml [source] to [destination] returning the
// fullfilled [destination] version (Config interface)
func MigrateDown(source interface{}, destination interface{}) utils.Config {
	var vStart, vFinish int
	if s, ok := source.(utils.Config); ok {
		_, vStart = findByVersion(s.V())
	}
	if d, ok := destination.(utils.Config); ok {
		_, vFinish = findByVersion(d.V())
	}

	var current = source
	for i := vStart; i > vFinish; i-- {
		next := reflect.New(reflect.TypeOf(configVersions[i-1].Config)).Interface()
		MigrateOne(current, next, configVersions[i].Down)
		current = next
	}
	return current.(utils.Config)
}
