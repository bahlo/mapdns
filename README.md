# mapdns [![CI](https://github.com/bahlo/mapdns/actions/workflows/ci.yml/badge.svg)](https://github.com/bahlo/mapdns/actions/workflows/ci.yml)

A smol DNS server (around 100 loc) that's configured with a static JSON file. Useful for split-dns.

## Install

### Go
```shell
go install github.com/bahlo/mapdns@v0.2.0
```

### Docker
```shell
docker pull ghcr.io/bahlo/mapdns:0.2.0
```

## Example
Create a `mapdns.json` in the same directory you're running the binary from, with content like this:
```json
{
	"foo.example.org.": {
		"A": "1.2.3.4",
		"AAAA": "::1"
	},
	"*.foo.example.org.": {
		"A": "1.2.3.4"
	}
}
```

Run the binary and start making requests!

## State of the project

It works and I use it in my home network for split-dns[^1]. 
There is little tests and no support for records other than `A` and `AAAA`. 
Please don't use this on a production system.

## Logging
Expose `MAPDNS_DEBUG=true` to get debug logs. Otherwise it will only log on 
errors. 

[^1]: I use Tailscale and configured it to search for my internal network domain
on the DNS server I configured here.
