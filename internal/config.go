package internal

import "os"

type Configuration struct {
	DatabaseFile string
}

func LoadEnvConfiguration() Configuration {
	return Configuration{
		DatabaseFile: os.Getenv("DB_FILE"),
	}
}