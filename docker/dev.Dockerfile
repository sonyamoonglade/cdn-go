FROM golang:1.18

WORKDIR /app

RUN apt update -y && \
   apt upgrade -y && \
   apt install -y \
   libvips-dev \
   build-essential \
   git && \
   go install github.com/githubnemo/CompileDaemon@latest

COPY . .

ENTRYPOINT CompileDaemon -polling -build="go build -o ./bin/cdn ./cmd/main.go" -command="./bin/cdn --debug=true --buckets-path=/data/buckets --config-path=./config.yaml"
