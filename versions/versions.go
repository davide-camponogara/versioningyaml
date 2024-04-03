package versions

type Migration map[string]func(c any) any

type ConfigVersion struct {
	Config Config
	Up     Migration
	Down   Migration
}

type Config interface {
	V() int
}

var ConfigVersions = []ConfigVersion{
	{
		ConfigV1{},
		nil,
		nil,
	},
	{
		ConfigV2{},
		MigrationUp,
		nil,
	},
	{
		ConfigV3{},
		MigrationUpV3,
		nil,
	},
}

var MigrationUp = Migration{
	"City": func(c any) any {
		conf := c.(ConfigV1)
		return conf.City + " " + conf.Street.Name
	},
}

var MigrationUpV3 = Migration{
	"TestV3": func(c any) any {
		conf := c.(ConfigV2)
		return float32(conf.Street.Field1)
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

func (ConfigV1) V() int {
	return 1
}

type ConfigV2 struct {
	Version int    `yaml:"version"`
	Street  Street `yaml:"street" comment:"$comm1"`
	City    string `yaml:"city" comment:"Test comment for city" lineComment:"test"`
}

func (ConfigV2) V() int {
	return 2
}

type ConfigV3 struct {
	Version int     `yaml:"version"`
	Street  Street  `yaml:"street" comment:"$comm1"`
	City    string  `yaml:"city" comment:"Test comment for city" lineComment:"test"`
	TestV3  float32 `yaml:"testv3" comment:"test v3"`
}

func (ConfigV3) V() int {
	return 3
}
