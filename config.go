package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// Config struct
type Config struct {
	DiscordConfig discordConfig     `toml:"discord"`
	DBConfig      databaseConfig    `toml:"database"`
	BotConfig   botConfig       `toml:"bot"`
}

// discordConfig struct
type discordConfig struct {
	Token   string `toml:"bot_token"`
}

// databaseConfig struct
type databaseConfig struct {
	DBFile             string `toml:"filename"`
}

// botConfig struct
type botConfig struct {
	// Command Prefix
	CP               string        `toml:"command_prefix"`
}

// ReadConfig function
func ReadConfig(path string) (config Config, err error) {

	var conf Config
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		fmt.Println(err)
		return conf, err
	}

	return conf, nil
}
