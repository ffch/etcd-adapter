package config

type server struct {
	Host string    `mapstructure:"host"`
	Port string    `mapstructure:"port"`
	TLS  serverTLS `mapstructure:"tls"`
}

type serverTLS struct {
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
}

type log struct {
	Level string `mapstructure:"level"`
}

type mysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type etcdConfig struct {
	Host     []string      `mapstructure:"host"`
	Prefix   string        `mapstructure:"prefix"`
	Timeout  int           `mapstructure:"timeout"`
	User     string        `mapstructure:"user"`
	Password string        `mapstructure:"password"`
	Tls      etcdTlsConfig `mapstructure:"tls"`
}

type etcdTlsConfig struct {
	CertFile string `mapstructure:"cert"`
	KeyFile  string `mapstructure:"key"`
	CaFile   string `mapstructure:"ca_file"`
	Verify   bool   `mapstructure:"verify"`
}

type datasource struct {
	Type  string      `mapstructure:"type"`
	MySQL mysqlConfig `mapstructure:"mysql"`
	Etcd  etcdConfig  `mapstructure:"etcd"`
}

type config struct {
	Server     server     `mapstructure:"server"`
	Log        log        `mapstructure:"log"`
	DataSource datasource `mapstructure:"datasource"`
}
