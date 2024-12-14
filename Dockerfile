FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY pkg ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o blastra-server ./pkg/main.go

FROM alpine:latest

USER 1000

WORKDIR /app

COPY --from=builder /app/blastra-server .

EXPOSE 8080

CMD ["./blastra-server"]
