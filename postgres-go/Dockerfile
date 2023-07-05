# # Dockerfile References: https://docs.docker.com/engine/reference/builder/

FROM golang:1.20-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./



# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

ENV DBURL=postgres://postgres:password@timescaledb/template1

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./main"]


