package main

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"io"
	"os"
)

/*
=== GLOBAL LOGGER CONFIGURATION ===
This file is used to configure the global logger for the application.
The global logger is used to log messages that are not specific to a particular package.
==================================
*/

type LogConfig struct {
	Level  log.Lvl
	Output io.Writer
}

var Logger = log.New("global")

func InitLog(cfg *LogConfig) error {
	if err := fixConfig(cfg); err != nil {
		return fmt.Errorf("log configuration error: %w", err)
	}

	Logger.SetLevel(cfg.Level)
	Logger.SetOutput(cfg.Output)
	return nil
}

func fixConfig(cfg *LogConfig) error {
	if cfg.Level == 0 {
		cfg.Level = log.INFO
	}
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return nil
}

func parseLogLevel(level string) log.Lvl {
	switch level {
	case "debug":
		return log.DEBUG
	case "info":
		return log.INFO
	case "warn":
		return log.WARN
	case "error":
		return log.ERROR
	case "off":
		return log.OFF
	default:
		panic("invalid log level")
	}
}

func parseLogOutput(output string) (io.Writer, func()) {
	switch output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic("failed to open log file: " + err.Error())
		}
		return file, func() { file.Close() }
	}
}
