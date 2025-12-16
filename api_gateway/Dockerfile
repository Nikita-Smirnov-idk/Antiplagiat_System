FROM golang:1.25.5-bookworm AS build

ARG GOPROXY="https://goproxy.io,direct"
ARG GOSUMDB="sum.golang.org"

WORKDIR /app

COPY services/storage_service/ ./services/storage_service/
COPY services/plagiarism_service/ ./services/plagiarism_service/
COPY api_gateway/go.mod api_gateway/go.sum ./api_gateway/

WORKDIR /app/api_gateway

RUN GONOSUMDB=* GOPROXY=${GOPROXY} GOSUMDB=${GOSUMDB} go mod download

COPY api_gateway/ .

RUN GONOSUMDB=* GOPROXY=${GOPROXY} GOSUMDB=${GOSUMDB} go build \
    -ldflags="-linkmode external -extldflags -static" \
    -tags netgo \
    -o /app/api-gateway ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /app/api-gateway /app/api-gateway

CMD ["/app/api-gateway"]