FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /server ./cmd/server

FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
