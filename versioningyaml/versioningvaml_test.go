package versioningyaml

import (
	"fmt"
	"testing"

	versions "github.com/davide-camponogara/versioningyaml/versions_test"
)

func TestMigration(t *testing.T) {
	SetConfigVersions(versions.ConfigVersions)
	SetLongComments(versions.LongComments)

	config, version, err := LoadConfigVersioned("../configv2.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Printf("config version %d: %#v", version, config)

	c, err := MigrateUp(config, &versions.ConfigV3{})
	if err != nil {
		panic(err)
	}
	configv2 := c.(*versions.ConfigV3)

	fmt.Printf("%#v", configv2)
	WriteYaml(*configv2, "../config_test_v2.yaml")
}
