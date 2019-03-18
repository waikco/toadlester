package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tsenart/vegeta/lib"

	"github.com/coocood/freecache"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/javking07/toadlester/conf"
	"github.com/javking07/toadlester/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// App ...
type App struct {
	AppServer  *http.Server
	AppClient  *http.Client
	AppStorage model.Storage
	AppRouter  *chi.Mux
	AppCache   *freecache.Cache
	AppLogger  *zerolog.Logger
	AppConfig  *conf.Config
	AppChans   map[string]chan struct{}
}

const timerChannel = "timerChannel"

func (a *App) RunApp() {
	// bootstrap
	log.Info().Msg("bootstrapping app")

	a.InitLogger()
	if a.AppLogger != nil {
		a.AppLogger.Info().Msgf("sleeping for %s to wait for dependencies", a.AppConfig.Sleep.String())
	}
	time.Sleep(*a.AppConfig.Sleep)

	a.InitCache()
	if err := a.InitDatabase(); err != nil {
		log.Fatal().Msgf("error bootstrapping database: %s", err.Error())
	}
	//a.InitClient()
	a.InitServer()

	// Start app by kicking off cron/ticker jobs and running server to take commands
	// todo set timer to accept jobs and adjustments to current jobs, from requests coming into server
	// todo add functionality to export metrics to influx

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	signal.Notify(gracefulStop, syscall.SIGKILL)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		a.AppLogger.Info().Msg("shutting down server")
		a.AppServer.Shutdown(context.Background())
		a.AppLogger.Info().Msg("shutting down timer")
		fmt.Println("Wait for 2 second to finish processing")
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	go a.InitTimer()

	go a.AppServer.ListenAndServe()

	select {
	case sig := <-gracefulStop:
		a.AppLogger.Info().Msgf("caught sig: %+v", sig)
		a.AppLogger.Info().Msg("Wait for 2 second to finish processing")
		a.AppLogger.Info().Msg("shutting down server")
		_ = a.AppServer.Shutdown(context.Background())
		for _, v := range a.AppChans {
			a.AppLogger.Info().Msgf("shutting down background processes: %s")
			close(v)
		}
		os.Exit(0)
	}

	//only run app is db connection present
	//if a.AppStorage != nil {
	//	//defer a.AppStorage // for graceful db shutdown
	//	a.AppServer.ListenAndServe()
	//}
}

func (a *App) InitChans(done chan struct{}) {
	appChans := make(map[string]chan struct{})
	appChans[timerChannel] = make(chan struct{})
	a.AppChans = appChans
}

func (a *App) InitTimer() {
	ticker := time.NewTicker(*a.AppConfig.Timer.Interval)

	a.AppLogger.Info().Msgf("initializing background process to run every: %s", *a.AppConfig.Timer.Interval)
	for {
		select {
		case <-a.AppChans[timerChannel]:
			a.AppLogger.Info().Msg("shutting down timer process")
			ticker.Stop()
			return

		case t := <-ticker.C:
			a.AppLogger.Info().Msgf("running job at: %s", t)
			cachedItems := a.AppCache.NewIterator()
			var tests []model.LoadTest
			a.AppLogger.Info().Msg("receiving tests from cache")
			// grab tests in cache
			for {
				if item := cachedItems.Next(); item == nil {
					break
				} else {
					var nextTest *model.LoadTest
					err := json.Unmarshal(item.Value, &nextTest)
					if err != nil {
						a.AppLogger.Error().Msgf("error unmarshalling test %v: %v", string(item.Key), err.Error())
						continue
					} else {
						tests = append(tests, *nextTest)
					}
				}
			}

			// run each test
			for _, test := range tests {
				a.AppLogger.Info().Msgf("running test: %+v", test)
				results, err := a.runTest(test)
				if err != nil {
					a.AppLogger.Error().Msg(err.Error())
				} else {
					a.AppLogger.Info().Msg(string(results))
					// todo : add test results to database
					err := a.AppStorage.Insert(test.Name, results)
					if err != nil {
						a.AppLogger.Error().Msg(err.Error())
					}
				}
			}
		}
	}

}

func (a *App) runTest(test model.LoadTest) ([]byte, error) {
	// set up test
	targetURL, _ := url.Parse(test.Url)
	targets := vegeta.NewStaticTargeter(vegeta.Target{
		Method: test.Method,
		URL:    targetURL.String(),
	})
	var duration time.Duration
	duration, err := time.ParseDuration(test.Duration)
	if err != nil {
		return nil, err
	}
	rate := vegeta.Rate{Freq: test.TPS, Per: duration}
	attacker := vegeta.NewAttacker()
	defer attacker.Stop()

	// run test
	var metrics vegeta.Metrics

	for res := range attacker.Attack(targets, rate, duration, test.Name) {
		metrics.Add(res)
	}
	metrics.Close()

	if jsonMetrics, err := json.Marshal(metrics); err != nil {
		return nil, err
	} else {
		return jsonMetrics, nil
	}
}

func (a *App) InitLogger() {
	var level string
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	log.Output(logger)

	// extract logging level from config is exists
	level = a.AppConfig.Logging.Level

	if level, err := zerolog.ParseLevel(level); err != nil {
		// if error parsing error log level, default to warn
		log.Warn().Msgf("error creating logger: %s", err.Error())
		logger.Level(zerolog.WarnLevel)
	} else {
		logger.Level(level)
	}

	a.AppLogger = &logger
	a.AppLogger.Info().Msgf("initializing logger to level `%s`", level)
}

// InitServer bootstraps app server
func (a *App) InitCache() {
	cacheSize := a.AppConfig.Cache.Size
	log.Info().Msgf("initializing cache with size of `%d` bytes", cacheSize)
	cache := freecache.NewCache(cacheSize)
	a.AppCache = cache
}

// InitDatabase bootstraps app storage
func (a *App) InitDatabase() error {
	db, err := model.BootstrapPostgres(a.AppConfig.Database)
	if err != nil {
		return err
	}
	a.AppStorage = db
	a.AppLogger.Info().Msgf("connected to database on port %v", a.AppConfig.Database.Port)
	return nil
}

// InitClient bootstraps app client
func (a *App) InitClient() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		},
		Timeout: 0,
	}
	a.AppClient = client
}

// InitServer bootstraps app server with handlers
func (a *App) InitServer() {

	a.AppRouter = chi.NewRouter()

	a.AppRouter.Use(middleware.RequestID)
	a.AppRouter.Use(middleware.RealIP)
	a.AppRouter.Use(middleware.Logger)
	a.AppRouter.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	a.AppRouter.Use(middleware.Timeout(60 * time.Second))

	// create actual routes
	a.AppRouter.Handle("/toadlester/v1/metrics", promhttp.Handler())
	a.AppRouter.Get("/toadlester/v1/health", a.Health)
	a.AppRouter.Post("/toadlester/v1/", a.PostTest)

	// Create server
	addr := fmt.Sprintf(":%s", a.AppConfig.Server.Port)
	a.AppServer = &http.Server{
		Addr:    addr,
		Handler: a.AppRouter,
	}

	if a.AppConfig.Server.TLS {
		cert, err := tls.LoadX509KeyPair(
			a.AppConfig.Server.Cert,
			a.AppConfig.Server.Key)

		if err != nil {
			log.Fatal().Msgf("Unable to load cert/key: %s", err)
		}

		cfg := &tls.Config{
			MinVersion:               tls.VersionTLS12,
			CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			InsecureSkipVerify: false,
			Certificates:       []tls.Certificate{cert},
		}
		cfg.BuildNameToCertificate()
		a.AppServer.TLSConfig = cfg
	}

	a.AppLogger.Info().Msgf("initialized routes and server on port %+v", a.AppServer)
}
