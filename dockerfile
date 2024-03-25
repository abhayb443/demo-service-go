# Use a base image with Go installed
FROM golang:1.17-alpine AS build

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o demo-service-go .

# Use a minimal base image for the final container
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built executable from the build container
COPY --from=build /app/demo-service-go .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./demo-service-go"]
