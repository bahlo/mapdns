package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

func buildLogger() (*zap.Logger, error) {
	if debug, _ := strconv.ParseBool(os.Getenv("MAPDNS_DEBUG")); debug {
		return zap.NewDevelopment()
	} else {
		return zap.NewProduction()
	}
}

func main() {
	logger, err := buildLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed build logger: %v", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v", err)
		}
	}()

	cfg, err := ReadConfig("mapdns.json")
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
		return
	}
	logger.Debug("Read config", zap.Any("config", cfg))

	srv := &dns.Server{Addr: ":53", Net: "udp"}
	srv.Handler = &Handler{logger: logger, cfg: cfg}

	logger.Info("Starting server", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Failed to set udp listener", zap.Error(err))
	}
}
