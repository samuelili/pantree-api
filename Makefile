.PHONY: build run start clean

# The name of the executable to be built.
BINARY_NAME := pantree-api

# Builds the Go application.
build:
	go build -o $(BINARY_NAME)

# Runs the Go application directly from source.
run:
	go run .

#Starts the Go application in release mode by running the built executable.
# It depends on the 'build' command to ensure the executable exists.
start: build
	GIN_MODE=release ./$(BINARY_NAME)

# Cleans up build artifacts.
clean:
	rm -f $(BINARY_NAME)