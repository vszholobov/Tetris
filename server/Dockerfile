ARG GO_VERSION=1.21.6
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /app

COPY . .
RUN go mod download

RUN GOOS=linux go build -o ./tetrisServer

# Multi-stage build, we just use plain alpine for the final image.
FROM alpine:latest
ENV GOPATH=/app

# Copy the binary from the first stage.
COPY --from=build ${GOPATH}/tetrisServer ./tetrisServer
RUN chmod u+x ./tetrisServer

# Set the run command.
CMD ["./tetrisServer"]