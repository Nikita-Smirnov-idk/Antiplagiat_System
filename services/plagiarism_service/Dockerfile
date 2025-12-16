FROM golang:1.25.5-bookworm AS build

WORKDIR /app

COPY storage_service/ ./storage_service/

COPY plagiarism_service/go.mod plagiarism_service/go.sum ./plagiarism_service/

WORKDIR /app/plagiarism_service

RUN go mod download

COPY plagiarism_service/ .

RUN go build \
    -ldflags="-linkmode external -extldflags -static" \
    -tags netgo \
    -o /app/plagiarism-service ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/plagiarism-service /app/plagiarism-service
COPY --from=build /app/plagiarism_service/migrations /app/migrations

CMD ["/app/plagiarism-service"]