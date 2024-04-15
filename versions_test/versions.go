//go:build !test

package versions_test

import "github.com/davide-camponogara/versioningyaml/utils"

// LongComments is a map containing long comments
// by convenction a reference to a long comment is denoted with a $ in form of the name
var LongComments = map[string]string{
	"$comm1": "Test commento numeor 1",
}

// ConfigVersions contains an ordered list of the versions of the yaml file
// every entry is a ConfigVersion object that contains:
//   - the Config struct
//   - the CustomMigration UP
//   - the CustomMigration DOWN
//
// the custom migrations are unes when one have to manipulate the fileds instead
// if a field doesn't change from the version before or is one is a new (standalone) field a CustomMigration is not needed
var ConfigVersions = []utils.ConfigVersion{
	{
		Config: ConfigV1{},
		Up:     nil,
		Down:   nil,
	},
	{
		Config: ConfigV2{},
		Up:     UpV2,
		Down:   nil,
	},
	{
		Config: ConfigV3{},
		Up:     UpV3,
		Down:   DownV3,
	},
}
