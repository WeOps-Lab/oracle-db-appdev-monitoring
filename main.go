// Copyright (c) 2021, 2025, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
// Portions Copyright (c) 2016 Seth Miller <seth@sethmiller.me>

package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	cversion "github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	_ "github.com/sijms/go-ora/v2"

	"github.com/alecthomas/kingpin/v2"

	// Required for debugging
	// _ "net/http/pprof"

	"github.com/oracle/oracle-db-appdev-monitoring/alertlog"
	"github.com/oracle/oracle-db-appdev-monitoring/collector"
	"github.com/oracle/oracle-db-appdev-monitoring/vault"
)

var (
	// Version will be set at build time.
	Version            = "0.0.0.dev"
	metricPath         = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics. (env: TELEMETRY_PATH)").Default(getEnv("TELEMETRY_PATH", "/metrics")).String()
	defaultFileMetrics = kingpin.Flag("default.metrics", "File with default metrics in a TOML file. (env: DEFAULT_METRICS)").Default(getEnv("DEFAULT_METRICS", "")).String()
	customMetrics      = kingpin.Flag("custom.metrics", "Comma separated list of file(s) that contain various custom metrics in a TOML format. (env: CUSTOM_METRICS)").Default(getEnv("CUSTOM_METRICS", "")).String()
	queryTimeout       = kingpin.Flag("query.timeout", "Query timeout (in seconds). (env: QUERY_TIMEOUT)").Default(getEnv("QUERY_TIMEOUT", "5")).Int()
	maxIdleConns       = kingpin.Flag("database.maxIdleConns", "Number of maximum idle connections in the connection pool. (env: DATABASE_MAXIDLECONNS)").Default(getEnv("DATABASE_MAXIDLECONNS", "10")).Int()
	maxOpenConns       = kingpin.Flag("database.maxOpenConns", "Number of maximum open connections in the connection pool. (env: DATABASE_MAXOPENCONNS)").Default(getEnv("DATABASE_MAXOPENCONNS", "10")).Int()
	poolIncrement      = kingpin.Flag("database.poolIncrement", "Connection increment when the connection pool reaches max capacity. (env: DATABASE_POOLINCREMENT)").Default(getEnv("DATABASE_POOLINCREMENT", "-1")).Int()
	poolMaxConnections = kingpin.Flag("database.poolMaxConnections", "Maximum number of connections in the connection pool. (env: DATABASE_POOLMAXCONNECTIONS)").Default(getEnv("DATABASE_POOLMAXCONNECTIONS", "-1")).Int()
	poolMinConnections = kingpin.Flag("database.poolMinConnections", "Minimum number of connections in the connection pool. (env: DATABASE_POOLMINCONNECTIONS)").Default(getEnv("DATABASE_POOLMINCONNECTIONS", "-1")).Int()
	scrapeInterval     = kingpin.Flag("scrape.interval", "Interval between each scrape. Default is to scrape on collect requests.").Default("0s").Duration()
	logDisable         = kingpin.Flag("log.disable", "Set to 1 to disable alert logs").Default("1").Int()
	logInterval        = kingpin.Flag("log.interval", "Interval between log updates (e.g. 5s).").Default("15s").Duration()
	logDestination     = kingpin.Flag("log.destination", "File to output the alert log to. (env: LOG_DESTINATION)").Default(getEnv("LOG_DESTINATION", "/log/alert.log")).String()
	toolkitFlags       = webflag.AddFlags(kingpin.CommandLine, ":9161")
	host               = kingpin.Flag("host", "Oracle database service ip or domain").Default("127.0.0.1").String()
	port               = kingpin.Flag("port", "Oracle database service port").Default("1521").String()
	serviceName        = os.Getenv("SERVICE_NAME")
	isDG               = kingpin.Flag("isDataGuard", "Whether this is a DataGuard").Default("false").Bool()
	isASM              = kingpin.Flag("isASM", "Whether this is a ASM").Default("false").Bool()
	isRAC              = kingpin.Flag("isRAC", "Whether this is a RAC").Default("false").Bool()
	isArchiveLog       = kingpin.Flag("isArchiveLog", "Whether to collect archiveLog metrics").Default("false").Bool()
	DSN                string
)

func main() {
	promLogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promLogConfig)
	kingpin.HelpFlag.Short('\n')
	kingpin.Version(version.Print("oracledb_exporter"))
	kingpin.Parse()
	logger := promslog.New(promLogConfig)
	user := os.Getenv("USER")
	password := os.Getenv("PASSWORD")
	connectString := os.Getenv("DB_CONNECT_STRING")
	//dbrole := os.Getenv("DB_ROLE")
	tnsadmin := os.Getenv("TNS_ADMIN")

	if connectString == "" {
		// 拼接DSN字符串
		DSN = fmt.Sprintf("oracle://%v:%v@%v:%s/%s", url.QueryEscape(user), url.QueryEscape(password), *host, *port, serviceName)
	}

	// externalAuth - Default to user/password but if no password is supplied then will automagically set to true
	externalAuth := false

	vaultID, useVault := os.LookupEnv("OCI_VAULT_ID")
	if useVault {

		logger.Info("OCI_VAULT_ID env var is present so using OCI Vault", "vaultOCID", vaultID)
		password = vault.GetVaultSecret(vaultID, os.Getenv("OCI_VAULT_SECRET_NAME"))
	}

	freeOSMemInterval, enableFree := os.LookupEnv("FREE_INTERVAL")
	if enableFree {
		logger.Info("FREE_INTERVAL env var is present, so will attempt to release OS memory", "free_interval", freeOSMemInterval)
	} else {
		logger.Info("FREE_INTERVAL end var is not present, will not periodically attempt to release memory")
	}

	restartInterval, enableRestart := os.LookupEnv("RESTART_INTERVAL")
	if enableRestart {
		logger.Info("RESTART_INTERVAL env var is present, so will restart my own process periodically", "restart_interval", restartInterval)
	} else {
		logger.Info("RESTART_INTERVAL env var is not present, so will not restart myself periodically")
	}

	config := &collector.Config{
		User:          user,
		Password:      password,
		ConnectString: connectString,
		DSN:           DSN,
		//DbRole:             dsn.AdminRole(dbrole),
		ConfigDir:          tnsadmin,
		ExternalAuth:       externalAuth,
		MaxOpenConns:       *maxOpenConns,
		MaxIdleConns:       *maxIdleConns,
		PoolIncrement:      *poolIncrement,
		PoolMaxConnections: *poolMaxConnections,
		PoolMinConnections: *poolMinConnections,
		CustomMetrics:      *customMetrics,
		QueryTimeout:       *queryTimeout,
		DefaultMetricsFile: *defaultFileMetrics,
		IsDG:               *isDG,
		IsASM:              *isASM,
		IsRAC:              *isRAC,
		IsArchiveLog:       *isArchiveLog,
	}
	exporter, err := collector.NewExporter(logger, config)
	if err != nil {
		logger.Error("unable to connect to DB", "error", err)
	}

	if *scrapeInterval != 0 {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go exporter.RunScheduledScrapes(ctx, *scrapeInterval)
	}

	prometheus.MustRegister(exporter)
	prometheus.MustRegister(cversion.NewCollector("oracledb_exporter"))
	prometheus.Unregister(collectors.NewGoCollector())

	logger.Info("Starting oracledb_exporter", "version", Version)
	logger.Info("Build context", "build", version.BuildContext())
	logger.Info("Collect from: ", "metricPath", *metricPath)

	opts := promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}
	http.Handle(*metricPath, promhttp.HandlerFor(prometheus.DefaultGatherer, opts))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<html><head><title>Oracle DB Exporter " + Version + "</title></head><body><h1>Oracle DB Exporter " + Version + "</h1><p><a href='" + *metricPath + "'>Metrics</a></p></body></html>"))
	})

	// start a ticker to cause rebirth
	if enableRestart {
		duration, err := time.ParseDuration(restartInterval)
		if err != nil {
			logger.Info("Could not parse RESTART_INTERVAL, so ignoring it")
		}
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		go func() {
			<-ticker.C
			logger.Info("Restarting the process...")
			executable, _ := os.Executable()
			execErr := syscall.Exec(executable, os.Args, os.Environ())
			if execErr != nil {
				panic(execErr)
			}
		}()
	}

	// start a ticker to free OS memory
	if enableFree {
		duration, err := time.ParseDuration(freeOSMemInterval)
		if err != nil {
			logger.Info("Could not parse FREE_INTERVAL, so ignoring it")
		}
		memTicker := time.NewTicker(duration)
		defer memTicker.Stop()

		go func() {
			for {
				<-memTicker.C
				logger.Info("attempting to free OS memory")
				debug.FreeOSMemory()
			}
		}()

	}

	// start the log exporter
	if *logDisable == 1 {
		logger.Info("log.disable set to 1, so will not export the alert logs")
	} else {
		logger.Info(fmt.Sprintf("Exporting alert logs to %s", *logDestination))
		logTicker := time.NewTicker(*logInterval)
		defer logTicker.Stop()

		go func() {
			for {
				<-logTicker.C
				logger.Debug("updating alert log")
				alertlog.UpdateLog(*logDestination, logger, exporter.GetDB())
			}
		}()
	}

	// start the main server thread
	server := &http.Server{}
	if err := web.ListenAndServe(server, toolkitFlags, logger); err != nil {
		logger.Error("Listening error", "error", err)
		os.Exit(1)
	}

}

// getEnv returns the value of an environment variable, or returns the provided fallback value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
