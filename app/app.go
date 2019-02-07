package app

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coocood/freecache"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/javking07/toadlester/conf"
	"github.com/javking07/toadlester/model"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tsenart/vegeta/lib"
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
}

type TestInfo struct {
	name     string
	url      string        `json:"url"`
	duration time.Duration `json:"duration"`
	tps      int           `json:"tps"`
}

func (a *App) RunApp() {
	// bootstrap
	log.Info().Msgf("bootstrapping app with the following config:/n %+V", *a.AppConfig)

	a.InitLogger()
	if a.AppLogger != nil {
		a.AppLogger.Info().Msgf("sleeping for %s to wait for dependencies", a.AppConfig.Sleep.String())
	}
	time.Sleep(*a.AppConfig.Sleep)

	a.InitRoutes()
	a.InitCache()
	if err := a.InitDatabase(); err != nil {
		log.Fatal().Msgf("error bootstrapping database: %s", err.Error())
	}
	a.InitClient()
	a.InitServer()

	// Start app by kicking off cron/ticker jobs and running server to take commands
	// todo set timer to accept jobs and adjustments to current jobs, from requests coming into server
	// todo add functionality to export metrics to influx

	go a.InitTimer()

	a.AppServer.ListenAndServe()

	// only run app is db connection present
	//if a.AppStorage != nil {
	//	defer a.AppStorage.Close() // for graceful db shutdown
	//	a.AppServer.ListenAndServe()
	//}
}

func (a *App) InitTimer() {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	ticker := time.NewTicker(*a.AppConfig.Timer.Interval)

	a.AppLogger.Info().Msgf("starting background process to run every: %s", *a.AppConfig.Timer.Interval)
	for {
		select {
		case sig := <-gracefulStop:
			a.AppLogger.Info().Msgf("caught sig: %+v", sig)
			a.AppLogger.Info().Msg("Wait for 2 second to finish processing")
			a.AppLogger.Info().Msg("shutting down server")
			_ = a.AppServer.Shutdown(context.Background())
			a.AppLogger.Info().Msg("shutting down background process")
			ticker.Stop()
			os.Exit(0)
		case t := <-ticker.C:
			a.AppLogger.Info().Msgf("Running job at: %s", t)
			runTest()
		}
	}

}

func runTest() {
	duration := 30 * time.Second
	targetURL, _ := url.Parse("https://app.waik.co/food-crawler/v1/health")
	targets := vegeta.Targets{
		{
			Method: "GET",
			URL:    targetURL,
		},
	}

	attacker := vegeta.NewAttacker()

	res := attacker.Attack(targets, 100, duration)

	results, err := vegeta.ReportJSON(res)

	if err != nil {
		log.Panic().Msg(err.Error())
	} else {

		fmt.Printf("Metrics: %+v\n", string(results))
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
	a.AppLogger.Info().Msgf("initialized logger to level `%s`", level)
}

// InitServer bootstraps app server
func (a *App) InitRoutes() {
	a.AppRouter = chi.NewRouter()
	// A good base middleware stack
	a.AppRouter.Use(middleware.RequestID)
	a.AppRouter.Use(middleware.RealIP)
	a.AppRouter.Use(middleware.Logger)
	a.AppRouter.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	a.AppRouter.Use(middleware.Timeout(60 * time.Second))

	// add actual api routes
	a.AppRouter.HandleFunc("/food-crawler/v1/health", a.Health)
	a.AppRouter.Handle("/food-food-crawler/v1/metrics", promhttp.Handler())
}

// InitServer bootstraps app server
func (a *App) InitCache() {
	cacheSize := a.AppConfig.Cache.Size
	log.Info().Msgf("initializing cache with size of `%d` bytes", cacheSize)
	cache := freecache.NewCache(cacheSize)
	a.AppCache = cache
}

// InitS bootstraps app server
func (a *App) InitDatabase() error {
	db, err := model.BootstrapPostgres(a.AppConfig.Database)
	if err != nil {
		return err
	}
	a.AppStorage = db
	return nil
}

// InitServer bootstraps app server
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

	router := chi.NewRouter()
	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(60 * time.Second))

	// create actual routes
	router.HandleFunc("/toadlester/v1/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, " a okay")
	})
	router.Handle("/toadlester/v1/metrics", promhttp.Handler())
	router.Post("/toadlester/v1/", func(writer http.ResponseWriter, request *http.Request) {

	})

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

	a.AppLogger.Info().Msgf("initialized server on port %+v", a.AppServer)

}
