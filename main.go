package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	adapter "github.com/axiomhq/axiom-go/adapters/zap"
	"github.com/miekg/dns"
	"go.uber.org/zap"
)

type Entry struct {
	A string
}

type Config map[string]Entry

func (c Config) Lookup(domain string) (Entry, bool) {
	entry, ok := c[domain]
	if !ok {
		// Check if we have a wildcard match
		for configDomain, entry := range c {
			if strings.HasPrefix(configDomain, "*.") && strings.HasSuffix(domain, configDomain[2:]) {
				return entry, true
			}
		}
	}

	return entry, ok
}

func ReadConfig(fileName string) (Config, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config: %w", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode config: %w", err)
	}

	return cfg, nil
}

type Handler struct {
	logger *zap.Logger
	cfg    Config
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := h.cfg.Lookup(domain)
		h.logger.Debug("Looked up domain", zap.String("domain", domain), zap.String("A", address.A))

		if ok && address.A != "" {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address.A),
			})
		}
	}

	if err := w.WriteMsg(&msg); err != nil {
		h.logger.Error("Failed to write message", zap.Error(err))
	}
}

func buildLogger() (*zap.Logger, error) {
	var logger *zap.Logger
	if debug, _ := strconv.ParseBool(os.Getenv("MAPDNS_DEBUG")); debug {
		// Debug, verbose logging
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	} else {
		// Prod
		if os.Getenv("AXIOM_TOKEN") != "" {
			// Axiom is set up, use their adapter
			fmt.Fprintln(os.Stderr, "Using Axiom adapter")
			core, err := adapter.New()
			if err != nil {
				return nil, fmt.Errorf("failed to create adapter: %w", err)
			}
			logger = zap.New(core)
		} else {
			// Use the default production logger
			var err error
			logger, err = zap.NewProduction()
			if err != nil {
				return nil, fmt.Errorf("failed to create logger: %w", err)
			}
		}
	}

	return logger, nil
}

func main() {
	logger, err := buildLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed build logger: %v", err)
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logs: %v", err)
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
