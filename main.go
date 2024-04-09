package main

import (
	"fmt"

	vyaml "github.com/davide-camponogara/versioningyaml/versioningyaml"
	versions "github.com/davide-camponogara/versioningyaml/versions_test"
)

func main() {
	vyaml.SetConfigVersions(versions.ConfigVersions)
	vyaml.SetLongComments(versions.LongComments)

	config, version, err := vyaml.LoadConfigVersioned("configv2.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("config version %d: %#v", version, config)

	c, err := vyaml.MigrateUp(config, &versions.ConfigV3{})
	configv2 := c.(*versions.ConfigV3)

	fmt.Printf("%#v", configv2)
	vyaml.WriteYaml(*configv2, "config_test_v2.yaml")

}
