package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/markor147/peverel/internal/log"
	"github.com/markor147/peverel/internal/tasks"
)

//go:embed assets/*
var assetsFS embed.FS

var data = &tasks.PsqlData{}

func main() {

	port, _ := strconv.Atoi(os.Getenv("SERVER_PORT"))
	logLevel := os.Getenv("LOG_LEVEL")
	logOutput := os.Getenv("LOG_OUTPUT")
	connStr := os.Getenv("DB_CONN_STRING")

	// log configuration
	parsedLogLevel := log.ParseLogLevel(logLevel)
	parsedLogOutput, closeFunc := log.ParseLogOutput(logOutput)
	if closeFunc != nil {
		defer closeFunc()
	}
	if err := log.InitLog(&log.Config{
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
	log.Logger.SetHeader("${time_rfc3339} ${short_file}:${line} ${level} ${message}")

	//data = NewMemoryData()
	data.Init(connStr, log.Logger)

	wgServer, err := StartServer(
		&Config{
			Port:       port,
			Logger:     log.Logger,
			FileSystem: assetsFS,
			RoutesRegister: func(e *Echo) {
				e.GET("/", func(c echo.Context) error {
					return c.Render(http.StatusOK, "layout", map[string]string{
						"Title":   "peverel - home",
						"Content": "home",
					})
				})
				e.GET("/add-task", func(c echo.Context) error {
					return c.Render(http.StatusOK, "layout", map[string]string{
						"Title":   "peverel - new task",
						"Content": "add-task",
					})
				})
				e.GET("/page/home", func(c echo.Context) error {
					return c.Render(http.StatusOK, "page/home", nil)
				})
				e.GET("/page/add-task", func(c echo.Context) error {
					return c.Render(http.StatusOK, "page/add-task", nil)
				})
				e.GET("empty-string", func(c echo.Context) error {
					return c.String(http.StatusOK, "")
				})
				e.GET("/forms/new-task", GetNewTaskForm)
				e.GET("/forms/edit-task", GetEditTaskForm)
				e.GET("/forms/new-group", GetNewGroupForm)
				e.GET("/groups", GetGroups)
				e.GET("/tasks", GetTasks)
				e.GET("/task/:id/next-time", GetTaskNextTime)
				e.POST("/task", PostTask)
				e.PUT("/task/:id", PutTask)
				e.GET("/modal/task-info", GetModalTaskInfo)
				e.GET("/task/:id/group/name", GetTaskGroupName)
				e.POST("/group", PostGroup)
				e.PUT("/task/:id/complete", PutTaskComplete)
				e.PUT("/task/:id/unassign", PutTaskUnassign)
				e.DELETE("/task/:id", DeleteTask)
				e.DELETE("/group/:id", DeleteGroup)
				e.POST("/tasks/mock", CreateMockTasks)
				e.PUT("/group/:id/assign", PutGroupAssignTask)
				e.GET("/tasks/count", GetTasksCount)
				e.GET("/modal/inactive", GetModalInactive)
			},
		},
	)
	if err != nil {
		log.Logger.Fatalf("Error starting server: %v", err)
	}
	defer log.Logger.Info("Server exited")

	wgServer.Wait()
}
