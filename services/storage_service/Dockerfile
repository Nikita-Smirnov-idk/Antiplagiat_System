FROM golang:1.25.5-bookworm AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build \
    -ldflags="-linkmode external -extldflags -static" \
    -tags netgo \
    -o storage_service ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/storage_service /app/storage_service
COPY --from=build /app/migrations /app/migrations

CMD ["/app/storage_service"]