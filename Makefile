APP_NAME=dot-sync

build:
	go build -o $(APP_NAME) .

run: build
	./$(APP_NAME)
