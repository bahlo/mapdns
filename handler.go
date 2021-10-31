package main

import (
	"net"

	"github.com/miekg/dns"
	"go.uber.org/zap"
)

const TTL = 60

type Handler struct {
	logger *zap.Logger
	cfg    Config
}

func (h *Handler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := dns.Msg{}
	msg.SetReply(r)

	msg.Authoritative = true
	domain := msg.Question[0].Name
	address, ok := h.cfg.Lookup(domain)

	if ok {
		h.logger.Debug("Looked up domain", zap.String("domain", domain), zap.String("A", address.A))
		switch r.Question[0].Qtype {
		case dns.TypeA:
			msg.Answer = append(msg.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: TTL},
				A:   net.ParseIP(address.A),
			})
		case dns.TypeAAAA:
			msg.Answer = append(msg.Answer, &dns.AAAA{
				Hdr:  dns.RR_Header{Name: domain, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: TTL},
				AAAA: net.ParseIP(address.AAAA),
			})
		}
	} else {
		h.logger.Debug("Could not find domain", zap.String("domain", domain))
		// We still want to return to not produce timeouts
	}

	if err := w.WriteMsg(&msg); err != nil {
		h.logger.Error("Failed to write message", zap.Error(err))
	}
}
