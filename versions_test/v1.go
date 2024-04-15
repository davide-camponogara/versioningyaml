//go:build !test

package versions_test

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

func (ConfigV1) V() int {
	return 1
}
