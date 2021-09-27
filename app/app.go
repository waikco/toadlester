package app

import (
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	uuid "github.com/satori/go.uuid"

	"github.com/go-chi/chi"
	"github.com/javking07/toadlester/conf"
	"github.com/javking07/toadlester/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// App ...
type App struct {
	Server   *http.Server
	Storage  model.Storage
	Router   *chi.Mux
	Logger   *zerolog.Logger
	Channels map[string]chan ChannelMessage
}

type ChannelMessage struct {
	message string
}

const timerChannel = "timerChannel"

// Bootstrap prepares app for run by setting things up based on provided config.
func (a *App) Bootstrap(c *conf.Config) {
	log.Info().Msg("bootstrapping app")

	var err error
	a.Logger, err = InitLogger(c)
	if err != nil {
		log.Fatal().Err(err)
	} else {
		a.Logger.Info().Msgf("sleeping for %s to wait for dependencies", c.Sleep.String())
	}
	time.Sleep(*c.Sleep)

	a.Logger.Info().Msgf("initializing database")
	a.Storage, err = InitDatabase(c)
	if err != nil {
		log.Fatal().Err(err).Msg("error preparing storage")
	}

	a.Channels = InitChans()
}

// RunApp starts app functionality and ensures a graceful shutdown.
func (a *App) RunApp(c *conf.Config) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGKILL)

	go a.InitTimer(c)
	select {
	case sig := <-gracefulStop:
		a.Logger.Info().Msgf("caught sig: %+v", sig)
		for _, v := range a.Channels {
			v <- struct{ message string }{sig.String()}
			a.Logger.Info().Msg("Wait for 1 second to finish processing")
			time.Sleep(time.Second)
		}
		os.Exit(0)
	}
}

// InitTimer kicks off the timer process intended to run in the background.
func (a *App) InitTimer(c *conf.Config) {
	// todo add functionality to export metrics to influx
	ticker := time.NewTicker(*c.Timer.Interval)

	if a.Logger != nil {
		a.Logger.Info().Msgf("initializing background process to run every: %s", *c.Timer.Interval)
	}

	for {
		select {
		case <-a.Channels[timerChannel]:
			a.Logger.Info().Msg("shutting down timer process")
			ticker.Stop()
			return

		case t := <-ticker.C:
			a.Logger.Info().Msgf("running job at: %s", t)
			// run each test
			for _, test := range c.Tests {
				a.Logger.Info().Msgf("running test for: %v", test.Name)
				results, err := a.RunTest(test.Name, *test.Duration, test.TPS, test.Target)
				if err != nil {
					a.Logger.Error().Msgf("error running test: %v", err)
					continue
				}

				data, err := json.Marshal(*results)
				if err != nil {
					a.Logger.Error().Msgf("error converting test results to json: %v", err)
					continue
				}
				if _, err := a.Storage.Insert(uuid.NewV4().String(), test.Name, data); err != nil {
					a.Logger.Error().Msgf("error inserting test results: %v", err)
					continue
				}
			}
		}
	}
}

func InitChans() map[string]chan ChannelMessage {
	appChans := make(map[string]chan ChannelMessage)
	appChans[timerChannel] = make(chan ChannelMessage)
	return appChans
}

// InitLogger returns a logger with the configured level
func InitLogger(c *conf.Config) (*zerolog.Logger, error) {
	var level string
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	log.Output(logger)

	// extract logging level from config is exists
	level = c.Logging.Level

	if level, err := zerolog.ParseLevel(level); err != nil {
		// if error parsing error log level, default to warn
		log.Warn().Msgf("error creating logger: %s", err.Error())
		logger.Level(zerolog.WarnLevel)
	} else {
		logger.Level(level)
		log.Info().Msgf("initializing logger to level `%s`", level)
	}
	return &logger, nil
}

// InitDatabase bootstraps database and returns app storage.
func InitDatabase(c *conf.Config) (model.Storage, error) {
	db, err := model.BootstrapPostgres(c.Database)
	if err != nil {
		return nil, err
	} else {
		log.Info().Msgf("database ssl status is %s", c.Database.SslMode)
	}
	return db, nil
}
