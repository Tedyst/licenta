FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY api /app/api
COPY cmd /app/cmd
COPY db /app/db
COPY email /app/email
COPY middleware /app/middleware
COPY models /app/models
COPY telemetry /app/telemetry
COPY templates /app/templates
COPY main.go /app

RUN go build -o /app/licenta .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/licenta /usr/local/bin/api
EXPOSE 5000

ENTRYPOINT ["/usr/local/bin/api", "serve"]