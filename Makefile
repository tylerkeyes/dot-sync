APP_NAME=dot-sync

build:
	go build -o $(APP_NAME) ./cmd

run: build
	./$(APP_NAME)
