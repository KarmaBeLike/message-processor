FROM golang:1.20
WORKDIR /message_service
COPY . . 

RUN go build -o message_service cmd/server/main.go
RUN chmod +x message_service
CMD ["./message_service"]

#docker-compose build



















# # First stage: build the Go binary
# FROM golang:1.20 as build

# WORKDIR /app

# # Copy go.mod and go.sum files
# COPY go.mod go.sum ./

# # Download dependencies
# RUN go mod download

# # Copy the source code
# COPY . ./

# # Build the Go application
# RUN go build -o /message_service cmd/server/main.go

# # Second stage: create a smaller image with the binary
# FROM gcr.io/distroless/base-debian10

# WORKDIR /

# # Copy the compiled binary from the build stage
# COPY --from=build /message_service /message_service

# # Expose the application port
# EXPOSE 8080

# # Run the compiled binary
# ENTRYPOINT ["/message_service"]
