package utils

type CustomMigration map[string]func(c any) any

type ConfigVersion struct {
	Config Config
	Up     CustomMigration
	Down   CustomMigration
}

type Config interface {
	V() int
}
