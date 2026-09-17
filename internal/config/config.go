package config

import "github.com/meesooqa/go-lgr"

type AppConfig struct {
	Logger lgr.Config `yaml:"logger"`
	Feeds  []Feed     `yaml:"feeds"`

	RawGenerator        string `yaml:"generator"`
	RawItunesOwnerName  string `yaml:"itunes_owner_name"`
	RawItunesOwnerEmail string `yaml:"itunes_owner_email"`
}

func (c *AppConfig) Generator() string {
	return c.RawGenerator
}

func (c *AppConfig) ItunesOwnerName() string {
	return c.RawItunesOwnerName
}

func (c *AppConfig) ItunesOwnerEmail() string {
	return c.RawItunesOwnerEmail
}
