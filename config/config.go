package config

import "os"

type Config struct {
	WebPort string // Port number for the web server TODO_PORT
	WebDir  string // Directory for the web server TODO_WEBDIR
	DBFile  string // Directory and file for the database TODO_DBFILE
	Pass    string // Password for web server TODO_PASS
}

// LoadCongig loads the configuration from environment variables
func LoadCongig() *Config {
	var config Config
	config.WebPort = os.Getenv("TODO_PORT")
	if config.WebPort == "" {
		config.WebPort = "7540"
	}
	config.WebDir = os.Getenv("TODO_WEBDIR")
	if config.WebDir == "" {
		config.WebDir = "./web"
	}
	config.DBFile = os.Getenv("TODO_DBFILE")
	if config.DBFile == "" {
		config.DBFile = "./scheduler.db"
	}
	config.Pass = os.Getenv("TODO_PASSWORD")
	return &config
}
