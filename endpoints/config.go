package abelana

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// AbelanaConfig contains all the information we need to run Abelana
type AbelanaConfig struct {
	AuthEmail         string
	ProjectID         string
	Bucket            string
	RedisPW           string
	Redis             string
	ServerKey         string
	AutoFollowers     []string
	Silhouette        string
	TimelineBatchSize int
	UploadRetries     int
	EnableBackdoor    bool
}

var config = mustLoadConfig("private/abelana-config.json")

func abelanaConfig() *AbelanaConfig {
	return config
}

func mustLoadConfig(path string) *AbelanaConfig {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("can't read Abelana config in file %q: %v", path, err)
	}
	var cfg AbelanaConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		log.Fatalf("error parsing Abelana config in file %q: %v", path, err)
	}
	return &cfg
}
