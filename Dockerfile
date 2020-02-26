FROM golang:1.12.7 as builder
WORKDIR /go/src/image-syncer
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make
RUN chmod +x ./image-syncer
ENTRYPOINT ["./image-syncer"]
CMD ["--config", "/etc/image-syncer/image-syncer.json"]
