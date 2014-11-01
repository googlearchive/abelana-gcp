package abelana

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

// AbelanaConfig contains all the information we need to run Abelana
type AbelanaConfig struct {
	AuthEmail         string
	ProjectID         string
	Bucket            string
	RedisPW           string
	Redis             string
	TimelineBatchSize int
	UploadRetries     int
	EnableBackdoor    bool
	EnableStubs       bool
}

var config struct {
	sync.Mutex
	c *AbelanaConfig
}

func abelanaConfig() *AbelanaConfig {
	config.Lock()
	defer config.Unlock()
	if config.c == nil {
		config.c = mustLoadConfig("private/abelana-config.json")
	}
	return config.c
}

// loadAbelanaConfig loads the configuration from the config file specified by path.
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
