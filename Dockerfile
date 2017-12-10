FROM golang:1.9-alpine
RUN apk add --no-cache git wget \
    && wget http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz \
    && tar -xzvf GeoLite2-City.tar.gz \
    && ls GeoLite2-City_20171205 \
    && mv GeoLite2-City_*/GeoLite2-City.mmdb /geoip.mmdb
WORKDIR /go/src/github.com/bah2830/Log-Exporter/
COPY . .
RUN go get -d -v ./... \
    && CGO_ENABLED=0 GOOS=linux go build -o log-exporter cmd/log-exporter.go


FROM scratch
COPY --from=0 /geoip.mmdb /app/geoip.mmdb
COPY --from=0 /go/src/github.com/bah2830/Log-Exporter/log-exporter /app/log-exporter
EXPOSE 9090
ENTRYPOINT ["/app/log-exporter"]
CMD ["-h"]
