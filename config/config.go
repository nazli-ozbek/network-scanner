package config

import (
	"log"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
)

var K = koanf.New(".")

func LoadConfig() {
	err := K.Load(file.Provider("config.json"), json.Parser())
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
}
