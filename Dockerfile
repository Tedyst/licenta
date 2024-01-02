FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum /app/
RUN go mod download

COPY api /app/api
COPY bruteforce /app/bruteforce
COPY cache /app/cache
COPY ci /app/ci
COPY cmd /app/cmd
COPY db /app/db
COPY docs /app/docs
COPY email /app/email
COPY extractors /app/extractors
COPY messages /app/messages
COPY models /app/models
COPY nvd /app/nvd
COPY rbac /app/rbac
COPY scanner /app/scanner
COPY tasks /app/tasks
COPY telemetry /app/telemetry
COPY templates /app/templates
COPY worker /app/worker
COPY main.go /app/

RUN go build -o /app/licenta .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /app/licenta /usr/local/bin/api
EXPOSE 5000

ENTRYPOINT ["/usr/local/bin/api", "servelocal"]