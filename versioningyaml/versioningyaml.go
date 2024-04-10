package versioningyaml

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/davide-camponogara/versioningyaml/utils"

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

// defaultVersion contains the default version of a .yaml where the field "version" is not present
var defaultVersion int = 1

// SetDefaultVersion setter for defaultVersion
//
// defaultVersion contains the default version of a .yaml where the field "version" is not present
func SetDefaultVersion(defVersion int) {
	defaultVersion = defVersion
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
func WriteYaml(data utils.Config, path string) error {
	wrapErr := func(err error) error {
		return fmt.Errorf("writing yaml: %w", err)
	}
	// Create a YAML nodes representation of the Address struct
	yamlObject, err := GenerateYAMLobject(data)
	if err != nil {
		return wrapErr(fmt.Errorf("error generating YAML: %w", err))
	}
	yamlBytes, err := yaml.Marshal(yamlObject)
	if err != nil {
		return wrapErr(fmt.Errorf("error marshalling YAML: %w", err))
	}

	// Write the YAML to a file
	file, err := os.Create(path)
	if err != nil {
		return wrapErr(fmt.Errorf("error creating file: %w", err))
	}
	defer file.Close()

	_, err = file.Write(yamlBytes)
	if err != nil {
		return wrapErr(fmt.Errorf("error writing YAML to file: %w", err))
	}
	fmt.Printf("%v file created successfully.", path)

	return nil
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
			fieldName = strings.ToLower(field.Name) // Use the field name as the key if yaml tag is empty
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
func LoadYAML(path string, object interface{}) error {
	wrapErr := func(err error) error {
		return fmt.Errorf("loading yaml: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return wrapErr(fmt.Errorf("error opening file: %w", err))
	}
	defer file.Close()

	// Read the content of the file
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(object); err != nil {
		return wrapErr(fmt.Errorf("error decoding YAML: %w", err))
	}
	return nil
}

// getVersion returns version of config file
func getVersion(path string) (int, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("getting version: %w", err)
	}
	var data interface{}
	err := LoadYAML(path, &data)
	if err != nil {
		return 0, wrapErr(err)
	}

	// Assert data to a map[string]interface{}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return 0, wrapErr(errors.New("invalid YAML format"))
	}

	// Handle different versions based on the "version" field
	vField, ok := dataMap["version"]
	if !ok {
		return defaultVersion, nil
	}

	version, ok := vField.(int)
	if !ok {
		return 0, wrapErr(errors.New("version field is not an integer"))
	}
	return version, nil
}

// LoadConfigVersioned loads a config file as a struct of the related version and returns also the version
func LoadConfigVersioned(path string) (utils.Config, int, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("loading config versioned: %w", err)
	}
	version, err := getVersion(path)
	if err != nil {
		return nil, 0, wrapErr(err)
	}

	// Get the appropriate struct type based on version
	configType, _ := findByVersion(version)
	if configType == nil {
		return nil, 0, wrapErr(errors.New("error finding version"))
	}

	// Load YAML into the appropriate struct type
	configValue := reflect.New(reflect.TypeOf(configType.Config)).Interface()
	err = LoadYAML(path, configValue)
	if err != nil {
		return nil, 0, wrapErr(err)
	}
	config := reflect.Indirect(reflect.ValueOf(configValue)).Interface()

	return config.(utils.Config), version, nil
}

// Migrate apply migration from config source to config destination objects
func MigrateOne(source interface{}, destination interface{}, migration utils.CustomMigration) error {
	/*wrapErr := func(err error) error {
		return fmt.Errorf("migrating one version: %w", err)
	}*/
	// Get the value of the destination struct
	destValue := reflect.ValueOf(destination).Elem()

	// Get the value of the source struct
	var sourceValue reflect.Value
	if reflect.ValueOf(source).Kind() == reflect.Ptr {
		sourceValue = reflect.ValueOf(source).Elem()
	} else {
		sourceValue = reflect.ValueOf(source)
	}

	for i := 0; i < destValue.NumField(); i++ {
		field := destValue.Type().Field(i)
		fieldName := field.Name

		// if field is of struct type
		if field.Type.Kind() == reflect.Struct {
			sourceStruct := sourceValue.FieldByName(fieldName)
			destStruct := destValue.FieldByName(fieldName)
			migrateStruct(sourceStruct, destStruct, fieldName, migration, sourceValue, destination.(utils.Config).V())

		} else { // if field is not of type struct
			migrateField(sourceValue, destValue, fieldName, migration, sourceValue, destination.(utils.Config).V())
		}
	}
	return nil
}

func migrateField(sourceValue reflect.Value, destValue reflect.Value, fieldPath string, migration utils.CustomMigration, sourceConfig reflect.Value, version int) {
	// get field name from dotted path
	spl := strings.Split(fieldPath, ".")
	fieldName := spl[len(spl)-1]

	if f, ok := migration[fieldPath]; ok {
		newValue := f(sourceConfig.Interface())
		destField := destValue.FieldByName(fieldName)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(reflect.ValueOf(newValue))
		}
	} else if fieldName == "Version" {
		destField := destValue.FieldByName(fieldName)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(reflect.ValueOf(version))
		}
	} else {
		sourceField := sourceValue.FieldByName(fieldName)
		if sourceField.IsValid() {
			destField := destValue.FieldByName(fieldName)
			if destField.IsValid() && destField.CanSet() {
				destField.Set(sourceField)
			}
		}
	}
}

func migrateStruct(sourceValue reflect.Value, destValue reflect.Value, structName string, migration utils.CustomMigration, sourceConfig reflect.Value, version int) {
	for i := 0; i < destValue.NumField(); i++ {
		field := destValue.Type().Field(i)
		fieldName := field.Name

		// if field is of struct type
		if field.Type.Kind() == reflect.Struct {
			sourceStruct := sourceValue.FieldByName(fieldName)
			destStruct := destValue.FieldByName(fieldName)
			migrateStruct(sourceStruct, destStruct, structName+"."+fieldName, migration, sourceConfig, version)
		} else { // if field is not of type struct
			migrateField(sourceValue, destValue, structName+"."+fieldName, migration, sourceConfig, version)
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
func MigrateUp(source interface{}, destination interface{}) (utils.Config, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("migrating up yaml: %w", err)
	}

	var vStart, vFinish int
	s, ok := source.(utils.Config)
	if !ok {
		return nil, wrapErr(errors.New("source doesn't implement Config interface"))
	}
	_, vStart = findByVersion(s.V())
	if vStart < 0 {
		return nil, wrapErr(errors.New("version not find in source"))
	}

	d, ok := destination.(utils.Config)
	if !ok {
		return nil, wrapErr(errors.New("destination doesn't implement Config interface"))
	}
	_, vFinish = findByVersion(d.V())
	if vFinish < 0 {
		return nil, wrapErr(errors.New("version not find in destination"))
	}

	var current = source
	for i := vStart; i < vFinish; i++ {
		next := reflect.New(reflect.TypeOf(configVersions[i+1].Config)).Interface()
		err := MigrateOne(current, next, configVersions[i+1].Up)
		if err != nil {
			return nil, wrapErr(err)
		}
		current = next
	}
	return current.(utils.Config), nil
}

// MigrateDown applies the DOWN migrations form yaml [source] to [destination] returning the
// fullfilled [destination] version (Config interface)
func MigrateDown(source interface{}, destination interface{}) (utils.Config, error) {
	wrapErr := func(err error) error {
		return fmt.Errorf("migrating up yaml: %w", err)
	}
	var vStart, vFinish int
	s, ok := source.(utils.Config)
	if !ok {
		return nil, wrapErr(errors.New("source doesn't implement Config interface"))
	}
	_, vStart = findByVersion(s.V())
	if vStart < 0 {
		return nil, wrapErr(errors.New("version not find in source"))
	}

	d, ok := destination.(utils.Config)
	if !ok {
		return nil, wrapErr(errors.New("destination doesn't implement Config interface"))
	}
	_, vFinish = findByVersion(d.V())
	if vFinish < 0 {
		return nil, wrapErr(errors.New("version not find in destination"))
	}

	var current = source
	for i := vStart; i > vFinish; i-- {
		next := reflect.New(reflect.TypeOf(configVersions[i-1].Config)).Interface()
		err := MigrateOne(current, next, configVersions[i].Down)
		if err != nil {
			return nil, wrapErr(err)
		}
		current = next
	}
	return current.(utils.Config), nil
}
