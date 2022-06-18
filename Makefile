.PHONY: build
build:
	go build -o ./botrunner -v ./app/bot/ 
run:
	make
	./botrunner


.DEFAULT_GOAL := build
