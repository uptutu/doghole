# Stage 1: Build the application
FROM golang:alpine AS builder

# Set necessary environment variables
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application source code
COPY . .

# Arguments for version information
ARG APP_VERSION=dev
ARG COMMIT_HASH=unknown
ARG BUILD_TIME=unknown

# Build the application
# Ensure the PKG_PATH matches your project structure for version variables
RUN go build -ldflags="-X doghole/cmd.Version=${APP_VERSION} -X doghole/cmd.CommitHash=${COMMIT_HASH} -X doghole/cmd.BuildTime=${BUILD_TIME}" -o /app/doghole main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/doghole /app/doghole

# Copy the configuration file
# Note: For production, it's often better to mount configurations
# or use environment variables instead of baking the config into the image.
COPY config.yaml /app/config.yaml

# Expose the port the application runs on (default is 8080)
EXPOSE 8080

# Define the entrypoint for the container
ENTRYPOINT ["/app/doghole"]

# Default command to run when the container starts
CMD ["server", "--config", "/app/config.yaml"]
