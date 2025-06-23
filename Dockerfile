FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o github-scanner ./cmd/github-scanner

FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/github-scanner ./github-scanner
COPY scanners.yaml ./scanners.yaml
ENTRYPOINT ["/app/github-scanner"]
CMD ["-h"]
