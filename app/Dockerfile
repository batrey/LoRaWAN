#Builder Stage 
FROM golang:1.19.0-alpine3.16 AS builder
WORKDIR /app
WORKDIR /go/src/github.com/postgres-go
COPY . .
RUN gp get -u githumb.com/lib/pq
RUN go build -0 main /app/main.go

#Run Stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
CMD ["/app/main"]