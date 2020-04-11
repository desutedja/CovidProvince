#BUILDER
FROM golang:1.13-alpine3.11 as BUILDER

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.* /

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY . /app

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o covid-api ./main.go

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the Pre-built binary file from the previous stage. Observe we also copied the .env file
COPY --from=builder /app/covid-api .
#COPY --from=builder /app/.env .

# Expose port 8181 to the outside world
EXPOSE 8181

#Command to run the executable
CMD ./covid-api