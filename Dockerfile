# build stage
FROM golang:1.25.4 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o worklogger ./cmd/server

# run stage
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app/worklogger /app/worklogger
COPY --from=build /app/frontend /app/frontend
COPY --from=build /app/worklogger.db /app/worklogger.db

EXPOSE 8080
ENTRYPOINT ["/app/worklogger"]
