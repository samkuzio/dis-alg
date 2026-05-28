# Stage 1: Build the statically linked Go binary
FROM golang:1.26.3-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the statically linked binary
# CGO_ENABLED=0 ensures static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags="-w -s" -o dis-alg ./cmd/dis-alg

# Stage 2: Create the distroless runtime image
FROM gcr.io/distroless/static-debian12

WORKDIR /

# Copy the compiled binary from the builder stage
COPY --from=builder /app/dis-alg /dis-alg

# Set the container to run as a non-root user (standard in distroless/static)
USER nonroot:nonroot

# The entrypoint is the binary itself.
# To run as a hub: docker run <image> hub
# To run as a terminal/spoke node: docker run <image> node <hub-address>
ENTRYPOINT ["/dis-alg"]
