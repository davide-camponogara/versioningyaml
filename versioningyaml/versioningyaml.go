package versioningyaml

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	yamlObject, err := GenerateYAMLobject(data, 0)
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
// gets level gor styling the black lines in comments
func GenerateYAMLobject(data interface{}, level int) (*yaml.Node, error) {
	// Get the type of the data
	dataType := reflect.TypeOf(data)

	// Create a new YAML node for the root
	rootNode := &yaml.Node{
		Kind: yaml.MappingNode,
	}

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

		// Create key node
		keyNode := &yaml.Node{
			Kind:        yaml.ScalarNode,
			Value:       fieldName,
			HeadComment: commentTag,
		}

		// ######## assign value #########

		var valueNode *yaml.Node
		var err error
		// If field is of type struct
		if reflect.ValueOf(data).Field(i).Type().Kind() == reflect.Struct {
			val := reflect.ValueOf(data).Field(i).Interface()
			// if object has Config method for formatting
			if t, ok := val.(interface{ Config() string }); ok {
				val = t.Config()
				valueNode = &yaml.Node{
					Kind:        yaml.ScalarNode,
					Value:       val.(string), // Get the field value from the struct
					LineComment: lineCommentTag,
				}
			} else {
				valueNode, err = GenerateYAMLobject(reflect.ValueOf(data).Field(i).Interface(), level+1)
			}
			// skip line
			if (i == 0 || reflect.ValueOf(data).Field(i-1).Type().Kind() != reflect.Struct) && level <= 1 {
				keyNode.HeadComment = "\n" + keyNode.HeadComment
			}
		} else if field.Type.Kind() == reflect.Ptr { // If field is of type pointer
			val := reflect.ValueOf(data).Field(i).Elem()
			// if is not valid write null
			if !val.IsValid() {
				valueNode = &yaml.Node{
					Kind:        yaml.ScalarNode,
					Value:       "null", // Get the field value from the struct
					LineComment: lineCommentTag,
				}
				// skip line
				if (i == 0 || reflect.ValueOf(data).Field(i-1).Type().Kind() != reflect.Struct) && level <= 1 {
					keyNode.HeadComment = "\n" + keyNode.HeadComment
				}
			} else { // else write value
				valueNode = &yaml.Node{
					Kind:        yaml.ScalarNode,
					Value:       fmt.Sprintf("%v", val), // Get the field value from the struct
					LineComment: lineCommentTag,
				}
			}
		} else { // else if field is a simple type
			var val any
			field := reflect.ValueOf(data).Field(i)
			if field.IsValid() {
				// Check if the field is a pointer
				if field.Kind() == reflect.Ptr && field.IsNil() {
					// Field is a nil pointer (likely representing null)
					val = reflect.Zero(field.Type().Elem()).Interface()
					val = fmt.Sprintf("%v", val)
				} else if field.Kind() == reflect.Interface && field.IsNil() {
					// Field is a nil interface{} (likely representing null)
					val = nil
					val = fmt.Sprintf("%v", val)
				} else {
					// Field is not nil, get its value
					val = field.Interface()
					// if object has Config method for formatting
					if t, ok := val.(interface{ Config() string }); ok {
						val = t.Config()
					}
					val = fmt.Sprintf("%v", val)
				}
			} else {
				// Field is not valid (zero value), handle it appropriately
				// For example, you might want to assign a default value to val
				val = reflect.Zero(field.Type()).Interface()
				val = fmt.Sprintf("%v", val)
			}
			// Create value node
			valueNode = &yaml.Node{
				Kind:        yaml.ScalarNode,
				Value:       val.(string), // Get the field value from the struct
				LineComment: lineCommentTag,
			}
		}
		if err != nil {
			return nil, err
		}

		// skip line
		if i == dataType.NumField()-1 {
			valueNode.FootComment = "\n" + valueNode.FootComment
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

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return wrapErr(fmt.Errorf("error opening file: %w", err))
	}

	err = yaml.Unmarshal(bytes, object)
	if err != nil {
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
	wrapErr := func(err error) error {
		return fmt.Errorf("migrating one version: %w", err)
	}
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
			err := migrateStruct(sourceStruct, destStruct, fieldName, migration, sourceValue, destination.(utils.Config).V())
			if err != nil {
				return wrapErr(err)
			}

		} else { // if field is not of type struct
			err := migrateField(sourceValue, destValue, fieldName, migration, sourceValue, destination.(utils.Config).V())
			if err != nil {
				return wrapErr(err)
			}
		}
	}
	return nil
}

func migrateField(sourceValue reflect.Value, destValue reflect.Value, fieldPath string, migration utils.CustomMigration, sourceConfig reflect.Value, version int) error {
	// get field name from dotted path
	spl := strings.Split(fieldPath, ".")
	fieldName := spl[len(spl)-1]

	if f, ok := migration[fieldPath]; ok {
		newValue := f(sourceConfig.Interface())
		destField := destValue.FieldByName(fieldName)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(reflect.ValueOf(newValue).Convert(destField.Type()))
		}
	} else if fieldName == "Version" {
		destField := destValue.FieldByName(fieldName)
		if destField.IsValid() && destField.CanSet() {
			destField.Set(reflect.ValueOf(version).Convert(destField.Type()))
		}
	} else {
		sourceField := sourceValue.FieldByName(fieldName)
		if sourceField.IsValid() {
			destField := destValue.FieldByName(fieldName)
			if destField.IsValid() && destField.CanSet() {
				destField.Set(sourceField.Convert(destField.Type()))
			}
		}
	}
	return nil
}

func migrateStruct(sourceValue reflect.Value, destValue reflect.Value, structName string, migration utils.CustomMigration, sourceConfig reflect.Value, version int) error {
	wrapErr := func(err error) error {
		return fmt.Errorf("migrating struct: %w", err)
	}

	for i := 0; i < destValue.NumField(); i++ {
		field := destValue.Type().Field(i)
		fieldName := field.Name

		// if field is of struct type
		if field.Type.Kind() == reflect.Struct {
			sourceStruct := sourceValue.FieldByName(fieldName)
			destStruct := destValue.FieldByName(fieldName)
			err := migrateStruct(sourceStruct, destStruct, structName+"."+fieldName, migration, sourceConfig, version)
			if err != nil {
				return wrapErr(err)
			}
		} else { // if field is not of type struct
			err := migrateField(sourceValue, destValue, structName+"."+fieldName, migration, sourceConfig, version)
			if err != nil {
				return wrapErr(err)
			}
		}
	}
	return nil
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
