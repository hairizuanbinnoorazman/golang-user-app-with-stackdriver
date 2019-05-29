package main

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"cloud.google.com/go/profiler"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
)

const ServiceName = "user-service"
const ServiceVersion = "0.0.1"

var SavedUsers = []user{
	{Name: "Bob", Gender: "Male"},
	{Name: "Lily", Gender: "Female"},
	{Name: "Unknown", Gender: "Gender"},
}

type user struct {
	Name   string `json:"name"`
	Gender string `json:"gender"`
}

type indexHandler struct {
	Logger *logrus.Logger
}

func (h indexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start Index Handler")
	defer h.Logger.Info("End Index handler")
	type respIndexHandler struct {
		Status string `json:"status"`
	}
	resp := respIndexHandler{Status: "Ok"}
	respOutput, _ := json.Marshal(resp)
	w.Write(respOutput)
}

type userHandler struct {
	Logger *logrus.Logger
}

func (h userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Info("Start user handler")
	defer h.Logger.Info("End user handler")

	type respUserHandler struct {
		Users []user `json:"user"`
	}

	time.Sleep(3 * time.Second)

	resp := respUserHandler{Users: SavedUsers}
	encoder := json.NewEncoder(w)
	encoder.Encode(resp)
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyLevel: "severity",
		},
	})

	// Profiler initialization, best done as early as possible.
	if err := profiler.Start(profiler.Config{
		Service:        ServiceName,
		ServiceVersion: ServiceVersion,
	}); err != nil {
		// TODO: Handle error.
		logger.Error("Unable to load profiler")
		logger.Error(string(debug.Stack()))
	}

	// Create and register a OpenCensus Stackdriver Trace exporter.
	exporter, err := stackdriver.NewExporter(stackdriver.Options{})
	if err != nil {
		logger.Error(err)
		logger.Error(string(debug.Stack()))
	}
	trace.RegisterExporter(exporter)

	logger.Info("Application Start Up")
	http.Handle("/", indexHandler{Logger: logger})
	http.Handle("/users", userHandler{Logger: logger})
	logger.Fatal(http.ListenAndServe("127.0.0.1:8888", &ochttp.Handler{}))
}
