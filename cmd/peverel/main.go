package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	data "github.com/markor147/peverel/internal/data"
	"github.com/markor147/peverel/internal/log"
)

//go:embed assets/*
var assetsFS embed.FS

func mustClone(t *template.Template) *template.Template {
	cl, err := t.Clone()
	if err != nil {
		log.Logger.Fatal(err)
	}
	return cl
}

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

	// Mux initialisation
	mux := http.NewServeMux()

	// Base layout
	baseTmpl := template.Must(template.ParseFS(assetsFS, "assets/tmpl/base.html"))

	// Static assets
	fsys, err := fs.Sub(assetsFS, "assets/static")
	if err != nil {
		log.Logger.Fatal(err)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fsys))))

	// Register simple pages
	for r, f := range map[string]string{
		"GET /settings":  "settings.html",
		"GET /tasks/new": "new-task.html",
	} {
		route := r
		file := f
		t := template.Must(mustClone(baseTmpl).ParseFS(assetsFS, "assets/tmpl/"+file))
		mux.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			if err := t.ExecuteTemplate(w, "base", nil); err != nil {
				log.Logger.Errorf("execute template %q: %v", file, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

	// Register home page
	{
		const file = "home.html"
		t := template.Must(mustClone(baseTmpl).ParseFS(assetsFS, "assets/tmpl/"+file))
		mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
			tasks, err := data.Tasks("", "", true)
			if err != nil {
				log.Logger.Errorf("get tasks: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			if err := t.ExecuteTemplate(w, "base", tasks); err != nil {
				log.Logger.Errorf("execute template %q: %v", file, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

	// Register edit task
	{
		const file = "edit-task.html"
		t := template.Must(mustClone(baseTmpl).ParseFS(assetsFS, "assets/tmpl/"+file))
		mux.HandleFunc("GET /tasks/{id}/edit", func(w http.ResponseWriter, r *http.Request) {
			idStr := r.PathValue("id")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Logger.Errorf("parse id %q: %v", idStr, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			task, err := data.GetTask(data.TaskId(id))
			if err != nil {
				log.Logger.Errorf("get task with id %d: %v", id, err)
				http.Error(w, err.Error(), http.StatusNotFound)
			}

			if err := t.ExecuteTemplate(w, "base", task); err != nil {
				log.Logger.Errorf("execute template %q: %v", file, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
	}

	// Init server
	port := os.Getenv("SERVER_PORT")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Run server
	go func() {
		log.Logger.Infof("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatalf("listen: %v\n", err)
		}
	}()

	// Trap SIGINT and SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Logger.Fatalf("server shutdown failed: %v", err)
	}

	log.Logger.Info("server exited")
}

/* DISMISSING
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

			// === Unused ===
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
wgServer.Wait()*/
