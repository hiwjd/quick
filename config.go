package quick

import "github.com/BurntSushi/toml"

type (
	// Config 配置
	Config struct {
		APIAddr     string `toml:"api_addr"`
		MysqlDSN    string `toml:"mysql_dsn"`
		EnableDBLog bool   `toml:"enable_db_log"`
		Log         Log    `toml:"log"`
		Redis       Redis  `toml:"redis"`
	}

	// Redis redis配置
	Redis struct {
		Addr     string `toml:"addr"`
		Password string `toml:"password"`
		DB       int    `toml:"db"`
	}

	// Log 日志配置
	Log struct {
		Level      string `toml:"level"`       // 日志级别
		Output     string `toml:"output"`      // 文件路径（例子：log/http.log）或者`stdout`
		MaxSize    int    `toml:"max_size"`    // 单个日志文件的大小上限，单位MB
		MaxBackups int    `toml:"max_backups"` // 最多保留几个日志文件
		MaxAge     int    `toml:"max_age"`     // 保留天数
		Compress   bool   `toml:"compress"`    // 是否压缩
	}
)

// MustLoadConfig 读取配置
func MustLoadConfig(fn string) Config {
	var config Config
	if _, err := toml.DecodeFile(fn, &config); err != nil {
		panic(err.Error())
	}
	return config
}
