package main

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	DispatcherAddr   string
	ConsolidatorAddr string
	FileServerAddr   string
}

func parseConfig(path string) Config {

	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		switch parts[0] {
		case "dispatcher":
			cfg.DispatcherAddr = parts[1] + ":" + parts[2]
		case "consolidator":
			cfg.ConsolidatorAddr = parts[1] + ":" + parts[2]
		case "fileserver":
			cfg.FileServerAddr = parts[1] + ":" + parts[2]
		}

	}

	return cfg
}
