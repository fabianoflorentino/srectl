# Build context: repository root.
#   docker build -f apps/api-a/Dockerfile -t srectl/api-a:dev .
FROM golang:1.27 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-a ./apps/api-a

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /out/api-a /usr/local/bin/api-a
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/api-a"]
