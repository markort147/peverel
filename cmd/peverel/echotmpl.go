package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/labstack/echo/v4"
)

type Config struct {
	Port           int
	Logger         *log.Logger
	FileSystem     fs.FS
	RoutesRegister func(e *Echo)
	CustomFuncs    FuncMap
}

type Echo = echo.Echo
type FuncMap = template.FuncMap

// StartServer initializes the echo server
// and registers all the endpoints
func StartServer(cfg *Config) (*sync.WaitGroup, error) {
	// create the server and configure the middleware
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${time_rfc3339} ${method}${uri} ${status}(${error}) ${latency_human} ${bytes_in}b ${bytes_out}b\n",
		Output: cfg.Logger.Output(),
	}))
	e.Logger.SetLevel(cfg.Logger.Level())
	e.Use(middleware.Recover())

	// serve index and register custom routes
	e.Renderer = newTemplateRenderer(cfg.FileSystem, "assets/tmpl/*", cfg.CustomFuncs)
	e.FileFS("/style.css", "assets/css/style.css", cfg.FileSystem)
	e.FileFS("/favicon.ico", "assets/ico/favicon.ico", cfg.FileSystem)
	if cfg.RoutesRegister != nil {
		cfg.RoutesRegister(e)
	}

	// intercept shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-quit
		cfg.Logger.Info("Shutting down the server")

		// Create a context with a timeout to allow for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt to gracefully shut down the server
		if err := e.Shutdown(ctx); err != nil {
			cfg.Logger.Error("Server forced to shutdown: ", err)
		}

		cfg.Logger.Info("Server exiting")
	}()

	// start the server
	go func() {
		// http
		if err := e.Start(fmt.Sprintf(":%d", cfg.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatal("Error starting the server: ", err)
		}
	}()

	return &wg, nil
}

type templateRenderer struct {
	tmpl *template.Template
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, _ echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

func newTemplateRenderer(fs fs.FS, path string, funcMap template.FuncMap) echo.Renderer {
	tmpl := template.New("templates")
	if funcMap != nil {
		tmpl.Funcs(funcMap)
	}
	return &templateRenderer{
		tmpl: template.Must(tmpl.ParseFS(fs, path)),
	}
}
