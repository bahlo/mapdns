package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Entry struct {
	A    string
	AAAA string
}

type Config map[string]Entry

func (c Config) Lookup(domain string) (Entry, bool) {
	entry, ok := c[domain]
	if !ok {
		// Check if we have a wildcard match
		for configDomain, entry := range c {
			if strings.HasPrefix(configDomain, "*.") && strings.HasSuffix(domain, configDomain[2:]) {
				return entry, true
			}
		}
	}

	return entry, ok
}

func ReadConfig(fileName string) (Config, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, nil
}
