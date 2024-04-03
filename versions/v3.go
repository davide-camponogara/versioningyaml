package versions

import "versioningyaml/utils"

type ConfigV3 struct {
	Version  int     `yaml:"version"`
	Street   Street  `yaml:"street" comment:"$comm1"`
	City     string  `yaml:"city" comment:"Test comment for city" lineComment:"test"`
	TestV3   float32 `yaml:"testv3" comment:"test v3"`
	TestV3_2 float32 `yaml:"testv3_2" comment:"test v3"`
}

func (ConfigV3) V() int {
	return 3
}

var UpV3 = utils.CustomMigration{
	"TestV3": func(c any) any {
		conf := c.(ConfigV2)
		return float32(conf.Street.Field1)
	},
}

var DownV3 = utils.CustomMigration{
	"City": func(c any) any {
		conf := c.(ConfigV3)
		return conf.Street.Name
	},
}
