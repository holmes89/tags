package internal

import "os"

type Configuration struct {
	DatabaseFile string
	BucketName   string
}

func LoadEnvConfiguration() Configuration {
	return Configuration{
		DatabaseFile: os.Getenv("DB_FILE"),
		BucketName:   os.Getenv("BUCKET_NAME"),
	}
}
