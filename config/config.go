package config

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"os"
)

type Config struct {
	Bedhock struct {
		Address string
	}
	Bot struct {
		DiscordLogURL string
		UserMap       map[string]string
		Token         string
		GuildID       string
		LogChannel    string
		LogCategory   string
		ChatCategory  string
		ChatChannel   string
		PlayingRoleID string
	}
}

func (c Config) ReverseDiscordUser(discordUsername string) string {
	for k, v := range c.Bot.UserMap {
		if v == discordUsername {
			return k
		}
	}
	return ""
}

func LoadConfig() (Config, error) {
	c := Config{}
	if _, err := os.Stat("config.toml"); err != nil {
		return c, err
	}
	data, err := ioutil.ReadFile("config.toml")
	if err != nil {
		return c, err
	}
	if err := toml.Unmarshal(data, &c); err != nil {
		return c, err
	}

	data, _ = toml.Marshal(c)
	//if err := ioutil.WriteFile("config.toml", data, 0644); err != nil {
	//	return c, err
	//}
	return c, nil
}
