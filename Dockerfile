FROM golang:1.27-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go tool templ generate \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/maidcafe ./cmd/server

FROM alpine:3.22
RUN adduser -D -H -u 10001 maidcafe
COPY --from=build /out/maidcafe /usr/local/bin/maidcafe
USER maidcafe
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/maidcafe"]
