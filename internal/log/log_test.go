package log

import (
	"bytes"
	"os"
	"testing"

	glog "github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

func TestInitLog(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		cfg := &Config{}
		err := InitLog(cfg)
		assert.NoError(t, err)
		assert.Equal(t, glog.INFO, Logger.Level())
		assert.Equal(t, cfg.Output, Logger.Output())
	})

	t.Run("custom config", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := &Config{
			Level:  glog.DEBUG,
			Output: &buf,
		}
		err := InitLog(cfg)
		assert.NoError(t, err)
		assert.Equal(t, glog.DEBUG, Logger.Level())
		assert.Equal(t, &buf, Logger.Output())
	})

	t.Run("nil config", func(t *testing.T) {
		err := InitLog(nil)
		assert.Error(t, err)
	})

}

func TestFixConfig(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		cfg := &Config{}
		err := fixConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, glog.INFO, cfg.Level)
		assert.NotNil(t, cfg.Output)
		//if assert.IsType(t, &os.File{}, cfg.Output) {
		//	asFile := cfg.Output.(*os.File)
		//	assert.Equal(t, os.Stdout.Name(), asFile.Name())
		//}
		assert.Equal(t, os.Stdout, cfg.Output)
	})

	t.Run("custom values", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := &Config{
			Level:  glog.DEBUG,
			Output: &buf,
		}
		err := fixConfig(cfg)
		assert.NoError(t, err)
		assert.Equal(t, glog.DEBUG, cfg.Level)
		assert.Equal(t, &buf, cfg.Output)
	})
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  glog.Lvl
	}{
		{
			name:  "debug",
			level: "debug",
			want:  glog.DEBUG,
		},
		{
			name:  "info",
			level: "info",
			want:  glog.INFO,
		},
		{
			name:  "warn",
			level: "warn",
			want:  glog.WARN,
		},
		{
			name:  "error",
			level: "error",
			want:  glog.ERROR,
		},
		{
			name:  "off",
			level: "off",
			want:  glog.OFF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLogLevel(tt.level)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("panic on invalid level", func(t *testing.T) {
		assert.Panics(t, func() {
			ParseLogLevel("invalid")
		})
	})
}

func TestParseLogOutput(t *testing.T) {
	t.Run("stdout", func(t *testing.T) {
		out, cleanup := ParseLogOutput("stdout")
		assert.Equal(t, out, os.Stdout)
		assert.Nil(t, cleanup)
	})

	t.Run("stderr", func(t *testing.T) {
		out, cleanup := ParseLogOutput("stderr")
		assert.Equal(t, out, os.Stderr)
		assert.Nil(t, cleanup)
	})

	t.Run("file", func(t *testing.T) {
		filename := "test.log"
		out, cleanup := ParseLogOutput(filename)
		assert.NotNil(t, out)
		assert.NotNil(t, cleanup)
		cleanup()
		_ = os.Remove(filename)
	})

	t.Run("panic on invalid file", func(t *testing.T) {
		assert.Panics(t, func() {
			ParseLogOutput("/invalid/path/test.log")
		})
	})
}
