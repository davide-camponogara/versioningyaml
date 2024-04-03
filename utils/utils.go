package utils

// CustomMigration it's a map with key the field of the config and value the associated function
// that takes a version of yaml and returns the value for the field
type CustomMigration map[string]func(c any) any

// ConfigVersion contains a Config struct and the UP and DOWN custom migrations
type ConfigVersion struct {
	Config Config
	Up     CustomMigration
	Down   CustomMigration
}

// Config it's the interface that requires to implement a V() method that expose the version of the yaml
type Config interface {
	V() int
}
