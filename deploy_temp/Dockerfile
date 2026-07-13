# --- Build stage ---
# Must be >= the "go" directive in go.mod (currently 1.25.0 — bumped by a
# transitive dependency of firebase.google.com/go/v4). golang:1.22-alpine
# here would fail "go mod tidy" with "go.mod requires go >= 1.25.0".
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum* ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /legaltech-server ./cmd/server

# --- Run stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /root/
COPY --from=builder /legaltech-server .

EXPOSE 8080

CMD ["./legaltech-server"]