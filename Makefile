.PHONY: build-image
build-image:
	docker build --tag mixz-robo .

.PHONY: run-dev
run-dev:
	export APP_ENV=dev && go run .
