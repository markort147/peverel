package log

import (
	"fmt"
	glog "github.com/labstack/gommon/log"
	"io"
	"os"
)

/*
=== GLOBAL LOGGER CONFIGURATION ===
This file is used to configure the global logger for the application.
The global logger is used to log messages that are not specific to a particular package.
==================================
*/

type Config struct {
	Level  glog.Lvl
	Output io.Writer
}

var Logger = glog.New("global")

func InitLog(cfg *Config) error {
	if err := fixConfig(cfg); err != nil {
		return fmt.Errorf("log configuration error: %w", err)
	}

	Logger.SetLevel(cfg.Level)
	Logger.SetOutput(cfg.Output)
	return nil
}

func fixConfig(cfg *Config) error {
	if cfg.Level == 0 {
		cfg.Level = glog.INFO
	}
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}
	return nil
}

func ParseLogLevel(level string) glog.Lvl {
	switch level {
	case "debug":
		return glog.DEBUG
	case "info":
		return glog.INFO
	case "warn":
		return glog.WARN
	case "error":
		return glog.ERROR
	case "off":
		return glog.OFF
	default:
		panic("invalid log level")
	}
}

func ParseLogOutput(output string) (io.Writer, func()) {
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
