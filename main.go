package main

import (
    "os"
	"log"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config represents the structure of the configuration file.
type Config struct {
	Paths []string `mapstructure:"paths"`
}

func main() {
	createDefaultConfig("config.yml")

	config, err := loadConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	watcher, err := createWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for _, path := range config.Paths {
		startEventListening(watcher, path)
	}

	blockMainRoutine()
}

func createDefaultConfig(configFile string) {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := Config{
			Paths: []string{"/tmp", "/opt"},
		}

		viper.SetDefault("paths", defaultConfig.Paths)

		viper.SetConfigFile(configFile)
		if err := viper.WriteConfigAs(configFile); err != nil {
			log.Fatal("Error creating default config:", err)
		}
		log.Println("Default configuration created at", configFile)
	}
}

func loadConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func createWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return watcher, nil
}

func startEventListening(watcher *fsnotify.Watcher, path string) {
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				handleEvent(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				handleError(err)
			}
		}
	}()
	err := watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
}

func handleEvent(event fsnotify.Event) {
	log.Println("Event:", event)
	if event.Has(fsnotify.Write) {
		log.Println("Modified file:", event.Name)
	}
}

func handleError(err error) {
	log.Println("Error:", err)
}

func blockMainRoutine() {
	<-make(chan struct{})
}
