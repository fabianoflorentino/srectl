# Build context: repository root.
#   docker build -f apps/api-b/Dockerfile -t srectl/api-b:dev .
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-b ./apps/api-b

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/api-b /usr/local/bin/api-b
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api-b"]
