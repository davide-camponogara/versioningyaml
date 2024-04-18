//go:build !test

package versions_test

import "github.com/davide-camponogara/versioningyaml/utils"

type ConfigV2 struct {
	Version int    `yaml:"version"`
	Street  Street `yaml:"street" comment:"$comm1"`
	City    string `yaml:"city" comment:"Test comment for city" lineComment:"test"`
	Test    map[int]bool
}

func (ConfigV2) V() int {
	return 2
}

var UpV2 = utils.CustomMigration{
	"City": func(c any) any {
		conf := c.(ConfigV1)
		return conf.City + " " + conf.Street.Name
	},
}
