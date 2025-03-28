package main

import (
	"embed"
	"flag"
	"fmt"
	glog "github.com/labstack/gommon/log"
	"github.com/markort147/gopkg/echotmpl"
	"github.com/markort147/gopkg/log"
	"github.com/markort147/gopkg/ymlcfg"
	"io"
	"os"
)

//go:embed assets/*
var assetsFS embed.FS

var data Data = &PsqlData{}

//var data Data = &MemoryData{}

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Log struct {
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
	} `yaml:"log"`
	Database struct {
		ConnStr string `yaml:"conn_string"`
	} `yaml:"database"`
}

func parseLogLevel(level string) glog.Lvl {
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

func main() {

	// parse config file path
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to the configuration file")
	flag.Parse()
	if configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// load configuration
	cfg, err := ymlcfg.ParseFile[Config](configPath)
	if err != nil {
		_, err1 := fmt.Fprintf(os.Stderr, "Error loading config: %v", err)
		if err1 != nil {
			panic(err1)
		}
		os.Exit(1)
	}
	logLevel := parseLogLevel(cfg.Log.Level)
	logOutput, closeFunc := parseLogOutput(cfg.Log.Output)
	if closeFunc != nil {
		defer closeFunc()
	}

	// log configuration
	if err = log.Init(&log.Config{
		Output: logOutput,
		Level:  logLevel,
	}); err != nil {
		_, err1 := fmt.Fprintf(os.Stderr, "Error init logger: %v", err)
		if err1 != nil {
			panic(err1)
		}
		os.Exit(1)
	}

	//data = NewMemoryData()
	data.Init(cfg)

	wgServer, err := echotmpl.StartServer(
		&echotmpl.Config{
			Port:          cfg.Server.Port,
			LogOutputPath: log.Logger.Output(),
			LogLevel:      logLevel,
			DefLogger:     log.Logger,
			FileSystem:    assetsFS,
			RoutesRegister: func(e *echotmpl.Echo) {
				e.GET("/forms/new-task", GetNewTaskForm)
				e.GET("/forms/new-group", GetNewGroupForm)
				e.GET("/groups", GetGroups)
				e.GET("/tasks", GetTasks)
				e.GET("/task/:id/next-time", GetTaskNextTime)
				e.POST("/task", PostTask)
				e.POST("/group", PostGroup)
				e.PUT("/task/:id/complete", PutTaskComplete)
				e.POST("/tasks/mock", CreateMockTasks)
				e.PUT("/group/:id/assign", PutGroupAssignTask)
			},
			//CustomFuncs: echotmpl.FuncMap{
			//	"NextTimeGroup": func(id GroupId) string {
			//		tasks := data.Relations[id]
			//		minNextDay, _ := time.Parse("2006-01-02", "9999-12-31")
			//		today := time.Now()
			//
			//		for _, taskId := range tasks {
			//			task := data.Tasks[taskId]
			//			nextDay := task.LastCompleted.AddDate(0, 0, task.Period)
			//			if nextDay.Before(minNextDay) {
			//				minNextDay = nextDay
			//			}
			//		}
			//
			//		todayStr := today.Format("20060102")
			//		nextDayStr := minNextDay.Format("20060102")
			//
			//		if todayStr == nextDayStr {
			//			return "today"
			//		}
			//
			//		todayTime, _ := time.Parse("20060102", todayStr)
			//		nextDayTime, _ := time.Parse("20060102", nextDayStr)
			//		diff := int(nextDayTime.Sub(todayTime).Hours() / 24)
			//		if diff < 0 {
			//			return fmt.Sprintf("%d days ago", -diff)
			//		}
			//		return fmt.Sprintf("%d days", diff)
			//	},
			//},
		},
	)
	if err != nil {
		log.Logger.Fatalf("Error starting server: %v", err)
	}
	defer log.Logger.Info("Server exited")

	wgServer.Wait()

}
