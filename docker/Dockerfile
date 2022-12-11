FROM golang:1.18-alpine AS builder
WORKDIR /build
RUN apk add --no-cache vips vips-dev alpine-sdk
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN GOOS=linux go build -o out/animakuro-cdn cmd/main.go

FROM alpine:3.16
WORKDIR /app
EXPOSE 9001
RUN apk add --no-cache curl dumb-init vips
COPY --from=builder /build/docker/start-cdn.sh .
COPY --from=builder /build/out/animakuro-cdn /usr/local/bin/
HEALTHCHECK --interval=15s --timeout=20s --start-period=20s --retries=3 CMD curl -f http://localhost:9001/api/health || exit 1
ENTRYPOINT [ "/usr/bin/dumb-init", "--" ]
CMD ["sh", "/app/start-cdn.sh"]
