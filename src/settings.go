package src

type TomlSettings struct {
	DB     DatabaseInfo `toml:"database"`
	System SystemInfo `toml:"system"`
}

type DatabaseInfo struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	Charset  string
	Timezone string
}

type SystemInfo struct {
	RoutineCount int `toml:"routine_count"`
}