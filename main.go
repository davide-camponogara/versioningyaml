package main

import (
	"fmt"

	vyaml "github.com/davide-camponogara/versioningyaml/versioningyaml"
	versions "github.com/davide-camponogara/versioningyaml/versions_test"
)

func main() {
	vyaml.SetConfigVersions(versions.ConfigVersions)
	vyaml.SetLongComments(versions.LongComments)

	config, version := vyaml.LoadConfigVersioned("configv3.yaml")
	fmt.Printf("config version %d: %#v", version, config)

	configv2 := vyaml.MigrateDown(&versions.ConfigV3{}, &versions.ConfigV2{}).(*versions.ConfigV2)

	fmt.Printf("%#v", configv2)
	vyaml.WriteYaml(*configv2, "config_test_v2.yaml")

}
