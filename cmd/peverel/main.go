package main

import (
	"embed"
	"fmt"
	"os"
	"strconv"

	data "github.com/markor147/peverel/internal/data"
	"github.com/markor147/peverel/internal/log"
)

//go:embed assets/*
var assetsFS embed.FS

func main() {
	// Log initialisation
	logLevel := os.Getenv("LOG_LEVEL")
	logOutput := os.Getenv("LOG_OUTPUT")
	logHeader := "${time_rfc3339} ${short_file}:${line} ${level} ${message}"
	closer, err := log.Init(logLevel, logOutput, logHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v", err)
		os.Exit(1)
	}
	if closer != nil {
		defer closer.Close()
	}

	// Data initialisation
	connStr := os.Getenv("DB_CONN_STRING")
	if err := data.Init(connStr); err != nil {
		log.Logger.Fatal(err)
	}

	// Server initialisation
	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {
		log.Logger.Fatalf("parse SERVER_PORT: %v", err)
	}
	wgServer, err := StartServer(
		&Config{
			Port:       port,
			FileSystem: assetsFS,
			RoutesRegister: func(e *Echo) {
				// Index
				e.GET("/", GetLayoutHome)

				// Pages
				e.GET("/page/home", GetPageHome)
				e.GET("/page/settings", GetPageSettings)
				e.GET("/page/add-task", GetPageAddTask)
				e.GET("/page/edit-task", GetPageEditTask)

				// Settings
				e.GET("/settings", GetLayoutSettings)
				e.GET("/settings/add-task", GetLayoutAddTask)
				e.GET("/settings/edit-task", GetLayoutEditTask)

				// Tasks
				e.GET("/tasks", GetTasks)

				// Task
				e.GET("/task/:id/next-time", GetTaskNextTime)
				e.POST("/task", PostTask)
				e.DELETE("/task/:id", DeleteTask)
				e.PUT("/task/:id", PutTask)
				e.PUT("/task/:id/complete", PutTaskComplete)

				/* === Unused === */
				// e.GET("/forms/new-task", GetNewTaskForm)
				// e.GET("/forms/edit-task", GetEditTaskForm)
				// e.GET("/forms/new-group", GetNewGroupForm)
				// e.GET("/groups", GetGroups)
				// e.GET("/modal/task-info", GetModalTaskInfo)
				// e.GET("/task/:id/group/name", GetTaskGroupName)
				// e.POST("/group", PostGroup)
				// e.PUT("/task/:id/unassign", PutTaskUnassign)
				// e.DELETE("/group/:id", DeleteGroup)
				// e.PUT("/group/:id/assign", PutGroupAssignTask)
				// e.GET("/tasks/count", GetTasksCount)
				// e.GET("/modal/inactive", GetModalInactive)
			},
		},
	)
	if err != nil {
		log.Logger.Fatalf("Error starting server: %v", err)
	}
	defer log.Logger.Info("Server exited")
	wgServer.Wait()

	os.Exit(0)
}
