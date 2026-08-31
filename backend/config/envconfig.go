package config

import "github.com/kelseyhightower/envconfig"

func mustProcess(v any) {
	if err := envconfig.Process("", v); err != nil {
		panic("config: " + err.Error())
	}
}
