package main

import (
	"context"
	"errors"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/holmes89/tags/internal"
	"github.com/holmes89/tags/internal/database"
	"github.com/holmes89/tags/internal/handlers/rest"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"net/http"
)

//Default values for application -> move to config?
const (
	defaultPort = ":8081"
	defaultCORS = "*"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	app := NewApp()
	app.Run()
	logrus.WithField("error", <-app.Done()).Error("terminated")
}

// NewApp will create new FX application which houses the server configuration and loading
func NewApp() *fx.App {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return fx.New(
		fx.Provide(
			internal.LoadEnvConfiguration,
			database.NewBoltConnection,
			database.NewKVStore,
			database.NewGraphDatabase,
			database.NewRepository,
			NewMux,
		),
		fx.Invoke(
			rest.NewResourceHandler,
			rest.NewTagHandler,
		),
		fx.Logger(
			logger,
		),
	)
}

// NewMux handler will create new routing layer and base http server
func NewMux(lc fx.Lifecycle) *mux.Router {
	logrus.Info("creating mux")

	router := mux.NewRouter()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{defaultCORS})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "PATCH", "OPTIONS", "DELETE"})
	cors := handlers.CORS(originsOk, headersOk, methodsOk)

	router.Use(cors)
	handler := (cors)(router)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logrus.Infof("starting server on %s", defaultPort)
			go func() {
				if err := http.ListenAndServe(defaultPort, handler); err != nil {
					logrus.WithError(err).Fatal("http server failure")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("stopping server")
			return errors.New("exited")
		},
	})

	return router
}
