FROM docker.io/golang:1.20rc2-alpine as BUILDER
RUN apk add git
WORKDIR /app
COPY . .
RUN go mod vendor
RUN go build -o http-echo .

FROM docker.io/alpine
RUN adduser -D -h /app 1000
WORKDIR /app
COPY --from=BUILDER /app/http-echo .
USER 1000
EXPOSE 8080
ENTRYPOINT ["/app/http-echo"]
