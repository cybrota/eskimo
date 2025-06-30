# ─────────────── Build your Go binary ───────────────
FROM --platform=linux/amd64 golang:1.23 AS builder

WORKDIR /app

# Install goreleaser for cross compilation
RUN curl -sfL https://github.com/goreleaser/goreleaser/releases/latest/download/goreleaser_Linux_x86_64.tar.gz \
    | tar -xz -C /usr/local/bin goreleaser

# Download deps
COPY go.mod go.sum ./
RUN go mod download

# Build using goreleaser
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 goreleaser build --clean --snapshot --single-target -o /app/eskimo

# ────────── Runtime with Python & Git and HomeBrew installed ─────────
FROM python:3.11-slim-bullseye AS runtime

# Install Git and Python
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    unzip \
    && rm -rf /var/lib/apt/lists/*


# ────────── Install Scanners ──────────
# 1. Semgrep
RUN pip3 install semgrep

# 2. Scharf
RUN curl -sf https://raw.githubusercontent.com/cybrota/scharf/refs/heads/main/install.sh | sh

# 3. Trivy
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.63.0

# Copy in your Go binary and config
WORKDIR /app
COPY --from=builder /app/eskimo .
COPY scanners.yaml .

# By default, run your Go CLI;
# but Python (and Git) are also available in the container shell if you need them.
ENTRYPOINT ["./eskimo"]
CMD ["-h"]
