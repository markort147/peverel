package main

import (
	"embed"
	"fmt"
	glog "github.com/labstack/gommon/log"
	"github.com/markort147/gopkg/echotmpl"
	"github.com/markort147/gopkg/log"
	"io"
	"os"
	"strconv"
)

//go:embed assets/*
var assetsFS embed.FS

// var data Data = &MemoryData{}
var data Data = &PsqlData{}

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

	port, _ := strconv.Atoi(os.Getenv("PEVEREL_PORT"))
	logLevel := os.Getenv("PEVEREL_LOG_LEVEL")
	logOutput := os.Getenv("PEVEREL_LOG_OUTPUT")
	connStr := os.Getenv("PEVEREL_DB_CONN_STRING")

	// log configuration
	parsedLogLevel := parseLogLevel(logLevel)
	parsedLogOutput, closeFunc := parseLogOutput(logOutput)
	if closeFunc != nil {
		defer closeFunc()
	}
	if err := log.Init(&log.Config{
		Output: parsedLogOutput,
		Level:  parsedLogLevel,
	}); err != nil {
		_, err1 := fmt.Fprintf(os.Stderr, "Error init logger: %v", err)
		if err1 != nil {
			panic(err1)
		}
		os.Exit(1)
	}
	//log.Logger.SetHeader(logHeader)
	log.Logger.SetHeader("${time_rfc3339} ${short_file}:${line} ${level} ${message}")

	//data = NewMemoryData()
	data.Init(connStr)

	wgServer, err := echotmpl.StartServer(
		&echotmpl.Config{
			Port:          port,
			LogOutputPath: log.Logger.Output(),
			LogLevel:      parsedLogLevel,
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
		},
	)
	if err != nil {
		log.Logger.Fatalf("Error starting server: %v", err)
	}
	defer log.Logger.Info("Server exited")

	wgServer.Wait()

}
