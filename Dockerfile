FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/subscription-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/bin/subscription-service .
COPY ./db/migrations/ ./db/migrations/
COPY ./db/seeds/ ./db/seeds/

EXPOSE 8080

CMD ["./subscription-service"]