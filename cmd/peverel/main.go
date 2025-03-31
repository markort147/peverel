package main

import (
	"embed"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"strconv"
)

//go:embed assets/*
var assetsFS embed.FS

// var data Data = &MemoryData{}
var data Data = &PsqlData{}

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
	if err := InitLog(&LogConfig{
		Output: parsedLogOutput,
		Level:  parsedLogLevel,
	}); err != nil {
		_, err1 := fmt.Fprintf(os.Stderr, "Error init logger: %v", err)
		if err1 != nil {
			panic(err1)
		}
		os.Exit(1)
	}
	//Logger.SetHeader(logHeader)
	Logger.SetHeader("${time_rfc3339} ${short_file}:${line} ${level} ${message}")

	//data = NewMemoryData()
	data.Init(connStr)

	wgServer, err := StartServer(
		&Config{
			Port:       port,
			Logger:     Logger,
			FileSystem: assetsFS,
			RoutesRegister: func(e *Echo) {
				e.GET("/forms/new-task", GetNewTaskForm)
				e.GET("/forms/new-group", GetNewGroupForm)
				e.GET("/groups", GetGroups)
				e.GET("/tasks", GetTasks)
				e.GET("/task/:id/next-time", GetTaskNextTime)
				e.POST("/task", PostTask)
				e.POST("/group", PostGroup)
				e.PUT("/task/:id/complete", PutTaskComplete)
				e.PUT("/task/:id/unassign", PutTaskUnassign)
				e.DELETE("/task/:id", DeleteTask)
				e.DELETE("/group/:id", DeleteGroup)
				e.POST("/tasks/mock", CreateMockTasks)
				e.PUT("/group/:id/assign", PutGroupAssignTask)
			},
		},
	)
	if err != nil {
		Logger.Fatalf("Error starting server: %v", err)
	}
	defer Logger.Info("Server exited")

	wgServer.Wait()
}

func DeleteTask(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteTask(TaskId(taskId)); err != nil {
		Logger.Errorf("Error deleting task: %v", err)
		return err
	}
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func DeleteGroup(c echo.Context) error {
	groupId, _ := strconv.Atoi(c.Param("id"))
	if err := data.DeleteGroup(GroupId(groupId)); err != nil {
		Logger.Errorf("Error deleting group: %v", err)
		return err
	}
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}

func PutTaskUnassign(c echo.Context) error {
	taskId, _ := strconv.Atoi(c.Param("id"))
	if err := data.UnassignTask(TaskId(taskId)); err != nil {
		Logger.Errorf("Error unassigning task: %v", err)
		return err
	}
	return c.Render(http.StatusOK, "groups", data.GetGroups())
}
