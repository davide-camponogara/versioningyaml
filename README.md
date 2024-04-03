# Versioning for yaml file like config

In order to add a new verison of a yaml (config) file one have to:
- Copy the old struct and paste it in a new file v##.go. Make the changes
- Add the new struct into the ConfigVersions map in versions.go
- [OPTIONAL]  Create functions for CustomMigration UP and DOWN into v##.go
- [OPTIONAL]  Add long comments to the "LongComments" map in versions.go associating a key preceded by "$"

The module exposes the functions:
- MigrateOne: to migrate (up or down) of only one version
- MigrateUp: to migrate UP of a certain number of versions
- MigrateDown to migrate DOWN of a certain number of versions




