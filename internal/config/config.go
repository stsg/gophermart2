package config

import (
	"flag"
	"sync"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

var once sync.Once

type config struct {
	RunAddress     string `env:"RUN_ADDRESS" envDefault:":8080"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	DatabaseURI    string `env:"DATABASE_URI"`
	SecretToken    string `env:"TOKEN_SIGN_KEY" envDefault:"The Little Man Who Wasn't There"`
}

var instance *config

func load() func() {
	return func() {
		var cfg config
		if err := env.Parse(&cfg); err != nil {
			zap.L().Fatal("load config: parse flags", zap.Error(err))
		}

		flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, `server address to listen on`)
		flag.StringVar(&cfg.AccrualAddress, "r", cfg.AccrualAddress, `accrual system address`)
		flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, `file location to store data in`)
		flag.Parse()

		instance = &cfg
	}
}

func New() {
	once.Do(load())
}

func Get() *config {
	if instance == nil {
		New()
	}
	return instance
}
