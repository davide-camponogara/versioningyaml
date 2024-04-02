package main

var configVersions = map[int]interface{}{
	1: ConfigV1{},
}

var migrationUp = map[string]func(interface{}) interface{}{
	"Version": func(c interface{}) interface{} {
		return 2
	},
	"City": func(c interface{}) interface{} {
		conf := c.(*ConfigV1)
		return conf.City + " " + conf.Street.Name
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
