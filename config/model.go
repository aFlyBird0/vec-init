package config

type config struct {
	*MysqlConfig
	*RedisConfig
	Str2VecConfigs []*Str2VecConfig
}

type MysqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Table    string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Database int
}

type Str2VecConfig struct {
	Field string
	Url   string
}
