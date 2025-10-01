package config

import (
    "log"
    "os"

    "github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
    Server struct {
        Port int `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
    } `yaml:"server"`

    Downloader struct {
        Workers      int    `yaml:"workers" env:"DOWNLOADER_WORKERS" env-default:"4"`
        TasksDir     string `yaml:"tasks_dir" env:"DOWNLOADER_TASKS_DIR" env-default:"tasks"`
        DownloadsDir string `yaml:"downloads_dir" env:"DOWNLOADER_DOWNLOADS_DIR" env-default:"downloads"`
    } `yaml:"downloader"`
}

func LoadConfig() *Config {
    var cfg Config
    configPath := "config.yaml"

    if _, err := os.Stat(configPath); err == nil {
        if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
            log.Fatalf("cannot read config: %s", err)
        }
    }

    if err := cleanenv.ReadEnv(&cfg); err != nil {
        log.Fatalf("cannot read env: %s", err)
    }

    return &cfg
}
