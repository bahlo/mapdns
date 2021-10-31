package main

import (
	"fmt"
	"net"
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandler_ServeDNS(t *testing.T) {
	cfg := Config{
		"example.org": Entry{
			A:    "127.0.0.1",
			AAAA: "::1",
		},
	}

	handler := Handler{
		logger: zap.L(),
		cfg:    cfg,
	}

	tests := []struct {
		Name           string
		Question       dns.Question
		ExpectedAnswer string
	}{
		{
			Name: "A example.org",
			Question: dns.Question{
				Name:  "example.org",
				Qtype: dns.TypeA,
			},
			ExpectedAnswer: fmt.Sprintf("example.org\t%d\tIN\tA\t127.0.0.1", TTL),
		},
		{
			Name: "AAAA example.org",
			Question: dns.Question{
				Name:  "example.org",
				Qtype: dns.TypeAAAA,
			},
			ExpectedAnswer: fmt.Sprintf("example.org\t%d\tIN\tAAAA\t::1", TTL),
		},
		{
			Name: "A foo.example.org (unknown domain)",
			Question: dns.Question{
				Name:  "unknown.example.org",
				Qtype: dns.TypeA,
			},
		},
		{
			Name: "MX example.org (unsupported type)",
			Question: dns.Question{
				Name:  "example.org",
				Qtype: dns.TypeMX,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w := &testResponseWriter{}
			handler.ServeDNS(w, &dns.Msg{
				Question: []dns.Question{test.Question},
			})
			require.NotNil(t, w.writtenMsg)
			if test.ExpectedAnswer != "" && assert.Len(t, w.writtenMsg.Answer, 1) {
				assert.Equal(t, test.ExpectedAnswer, w.writtenMsg.Answer[0].String())
			}
		})
	}
}

type testResponseWriter struct {
	writtenMsg *dns.Msg
}

func (w *testResponseWriter) LocalAddr() net.Addr  { panic("unimplemented") }
func (w *testResponseWriter) RemoteAddr() net.Addr { panic("unimplemented") }
func (w *testResponseWriter) WriteMsg(msg *dns.Msg) error {
	w.writtenMsg = msg
	return nil
}
func (w *testResponseWriter) Write([]byte) (int, error) { panic("unimplemented") }
func (w *testResponseWriter) Close() error              { panic("unimplemented") }
func (w *testResponseWriter) TsigStatus() error         { panic("unimplemented") }
func (w *testResponseWriter) TsigTimersOnly(bool)       { panic("unimplemented") }
func (w *testResponseWriter) Hijack()                   { panic("unimplemented") }
