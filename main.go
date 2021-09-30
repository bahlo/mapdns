package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
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
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

type Handler struct {
	cfg Config
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		msg.Authoritative = true
		domain := msg.Question[0].Name
		address, ok := h.cfg.Lookup(domain)
		if ok && address.A != "" {
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A:   net.ParseIP(address.A),
			})
		}
	}
	w.WriteMsg(&msg)
}

func main() {
	cfg, err := ReadConfig("mapdns.json")
	if err != nil {
		log.Fatal(err)
	}

	srv := &dns.Server{Addr: ":53", Net: "udp"}
	srv.Handler = &Handler{cfg: cfg}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to set udp listener %s\n", err.Error())
	}
}
