# Stage 1: Build the Go application
FROM golang:alpine AS builder

WORKDIR /app
COPY . /app
WORKDIR /app/cmd/app
RUN go mod download
RUN CGO_ENABLED=0 go build -o main .

# Stage 2: Create a minimal image with the Go application and ffmpeg
FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/app/main .
RUN apk add --no-cache ffmpeg

ENV PORT=8080

# Define volume mount points
VOLUME ["/config"]

CMD ["./main"]