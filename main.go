package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

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

func buildLogger() *zap.Logger {
	var (
		logger *zap.Logger
		err    error
	)

	if debug, _ := strconv.ParseBool(os.Getenv("MAPDNS_DEBUG")); debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"error", "message": "Failed to set up logger", "error": %q}`, err)
		os.Exit(1)
	}

	return logger
}

func main() {
	logger := buildLogger()
	defer logger.Sync()

	cfg, err := ReadConfig("mapdns.json")
	if err != nil {
		logger.Error("Failed to read config", zap.Error(err))
	}
	logger.Debug("Read config", zap.Any("config", cfg))

	srv := &dns.Server{Addr: ":53", Net: "udp"}
	srv.Handler = &Handler{logger: logger, cfg: cfg}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("Failed to set udp listener", zap.Error(err))
	}
}
