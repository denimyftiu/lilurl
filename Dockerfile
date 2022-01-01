# Build stage
FROM golang:1.17-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o shortner ./cmd/shortner 

# Move binary to new container
FROM scratch

COPY --from=builder /app/shortner /shortner

EXPOSE 8080

CMD ["/shortner"]
