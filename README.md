# mapdns [![CI](https://github.com/bahlo/mapdns/actions/workflows/ci.yml/badge.svg)](https://github.com/bahlo/mapdns/actions/workflows/ci.yml)

A DNS server that's configured with a static JSON file.

## Example
Create a `mapdns.json` in the same directory you're running the binary from, with content like this:
```json
{
	"foo.example.org.": {
		"A": "1.2.3.4"
	},
	"*.foo.example.org.": {
		"A": "1.2.3.4"
	}
}
```

Run the binary and start making requests!
