package log

import (
	"fmt"
	"io"
	"os"

	glog "github.com/labstack/gommon/log"
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

func Init(levelStr, outputStr, headerStr string) (io.Closer, error) {
	// parse level
	var lvl glog.Lvl
	switch levelStr {
	case "debug":
		lvl = glog.DEBUG
	case "info":
		lvl = glog.INFO
	case "warn":
		lvl = glog.WARN
	case "error":
		lvl = glog.ERROR
	case "off":
		lvl = glog.OFF
	default:
		return nil, fmt.Errorf("invalid log level %q", levelStr)
	}

	// parse output
	var w io.Writer
	var c io.Closer
	switch outputStr {
	case "stdout":
		w = os.Stdout
	case "stderr":
		w = os.Stderr
	default:
		f, err := os.OpenFile(outputStr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file %q: %w", outputStr, err)
		}
		w = f
		c = f
	}

	Logger.SetLevel(lvl)
	Logger.SetOutput(w)
	Logger.SetHeader(headerStr)

	return c, nil
}
