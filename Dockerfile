##########################################
# Build golang container
##########################################
FROM golang:1.22.2-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Setup the working directory
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Build the bbox tool
COPY . .
RUN go build -o bbox

##########################################
# Runtime container
##########################################
FROM debian:bookworm
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/bbox /usr/local/bin/bbox
WORKDIR /usr/local/bin
ENTRYPOINT ["./bbox"]  
