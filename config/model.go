package config

type config struct {
	*ServerConfig
	*MysqlConfig
	*RedisConfig
	Str2VecConfigs []*Str2VecConfig
	*ConcurrencyConfig
}

type ServerConfig struct {
	Host      string
	Port      int
	VectorDir string
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

type ConcurrencyConfig struct {
	PageSize        int
	PatentPoolSize  int
	VectorPoolSize  int
	QueryWorkerSize int
}
