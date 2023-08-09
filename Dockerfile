FROM docker.io/golang:1.20rc2-alpine as BUILDER
RUN apk add git
WORKDIR /app
COPY . .
RUN go mod vendor
RUN go build -o traefik-opa-proxy .

FROM docker.io/alpine
RUN adduser -D -h /app 1000
WORKDIR /app
COPY --from=BUILDER /app/traefik-opa-proxy .
USER 1000
EXPOSE 8182
ENTRYPOINT ["/app/traefik-opa-proxy"]
