# Stage 1: Build the application
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

ARG TARGETPLATFORM

RUN echo "I am running on $BUILDPLATFORM, building for $TARGETPLATFORM"

WORKDIR /app

# Copy go.mod and go.sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application
# Adjust the output path and main package path if necessary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${TARGETPLATFORM} go build -a -installsuffix cgo -o /app/yanm_app ./cmd/main.go

# Stage 2: Create the runtime image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/yanm_app .

# Copy the configuration file
# This assumes config.yml is in the root of your project context
COPY config.yml .

# Expose any necessary ports if your application is a server (optional)
# For example: EXPOSE 8080
EXPOSE 8090

# Command to run the application
# The application will look for config.yml in the current working directory
ENTRYPOINT ["/app/yanm_app"]
