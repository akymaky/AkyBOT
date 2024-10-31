package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

var _config *Configuration

type overrides []struct {
	Match     string
	ChannelId int64 `yaml:"channel_id"`
	RoleId    int64 `yaml:"role_id"`
	Embed     discordEmbed
}

type discordEmbed struct {
	Title  string
	Author string
}

type Configuration struct {
	DiscordToken string `yaml:"discord_token"`
	Moodle       struct {
		Url       string
		Token     string
		ChannelId int64 `yaml:"channel_id"`
		Embed     discordEmbed
		Courses   []struct {
			Name      string
			Embed     discordEmbed
			ForumId   int   `yaml:"forum_id"`
			RoleId    int64 `yaml:"role_id"`
			ChannelId int64 `yaml:"channel_id"`
			Overrides overrides
		}
	}
	RSS []struct {
		Url        string
		ChannelId  int64 `yaml:"channel_id"`
		RoleId     int64 `yaml:"role_id"`
		Department string
		Overrides  overrides
	}
}

func LoadConfig() {
	file, err := os.Open("config.yml")
	if err != nil {
		panic("config.yml not found!")
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(&_config)
	if err != nil {
		panic(err)
	}
}

func Get() *Configuration {
	c := *_config
	return &c
}
