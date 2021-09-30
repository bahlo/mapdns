FROM golang:alpine AS builder
ADD . /go/src/github.com/bahlo/mapdns
WORKDIR /go/src/github.com/bahlo/mapdns
RUN go build -o /usr/bin/mapdns .

FROM alpine
COPY --from=builder /usr/bin/mapdns /usr/bin/mapdns
WORKDIR /opt/mapdns
ENTRYPOINT ["/usr/bin/mapdns"]
