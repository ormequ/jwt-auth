package config

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

const (
	EnvDebug   = gin.DebugMode
	EnvRelease = gin.ReleaseMode
)

type Config struct {
	Env            string `env:"ENV" env-default:"release"`
	MongoConn      string `env:"MONGO_CONN" env-required:"true"`
	MongoDB        string `env:"MONGO_DB" env-required:"true"`
	HTTPAddr       string `env:"HTTP_ADDR" env-default:":8888"`
	BCryptCost     int    `env:"BCRYPT_COST" env-default:"10"`
	AccessSecret   string `env:"ACCESS_SECRET_KEY" env-required:"true"`
	RefreshSecret  string `env:"REFRESH_SECRET_KEY" env-required:"true"`
	AccessExpires  int    `env:"ACCESS_EXPIRES" env-default:"300"`
	RefreshExpires int    `env:"REFRESH_EXPIRES" env-default:"2592000"` // default - 30 days
}

func MustLoad() Config {
	path := os.Getenv("CONFIG_PATH")
	var cfg Config
	var err error
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			panic(fmt.Sprintf("config file does not exist: %s", path))
		}
		err = cleanenv.ReadConfig(path, &cfg)
	} else {
		err = cleanenv.ReadEnv(&cfg)
	}
	if err != nil {
		panic(fmt.Sprintf("cannot read config: %s", err))
	}

	return cfg
}
