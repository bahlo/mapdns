FROM golang:alpine AS builder
ADD . /go/src/github.com/bahlo/mapdns
WORKDIR /go/src/github.com/bahlo/mapdns
RUN go build -o /usr/bin/mapdns .

FROM alpine
LABEL org.opencontainers.image.source=https://github.com/bahlo/mapdns
COPY --from=builder /usr/bin/mapdns /usr/bin/mapdns
WORKDIR /opt/mapdns
EXPOSE 53/udp
ENTRYPOINT ["/usr/bin/mapdns"]
