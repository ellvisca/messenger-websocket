FROM golang:alpine AS builder

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /app
WORKDIR /app

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
COPY /public/* ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o main .

# Build a small image
FROM alpine:latest

WORKDIR /root/
RUN mkdir public

ENV REDIS_URL=redis-19978.c228.us-central1-1.gce.cloud.redislabs.com:19978 \
    REDIS_PASSWORD=Uv7u3u547C7hbkedhiBLVa8BGJMwLV8m

# Expose port 8080 to the outside world
EXPOSE 8080

COPY --from=builder /app/main .
COPY --from=builder /app/public .

# Command to run
ENTRYPOINT ["./main"]